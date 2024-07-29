//go:build integration

package tfapstra_test

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testcheck "github.com/Juniper/terraform-provider-apstra/apstra/test_check_funcs"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

func systemIds(ctx context.Context, t *testing.T, bp *apstra.TwoStageL3ClosClient, role string) []string {
	query := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal(role)},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	var result struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	err := query.Do(ctx, &result)
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, len(result.Items))
	for i, item := range result.Items {
		ids[i] = item.System.Id
	}

	return ids
}

func stringPtrOrNull(in *string) string {
	if in == nil {
		return "null"
	}
	return `"` + *in + `"`
}

func stringOrNull(in string) string {
	if in == "" {
		return "null"
	}
	return `"` + in + `"`
}

func stringMapOrNull(in map[string]string, depth int) string {
	if in == nil {
		return "null"
	}

	if len(in) == 0 {
		return "{}"
	}

	whitespace := strings.Repeat("  ", depth)

	var longestKey int
	for k := range in {
		keyLen := len(k)
		if keyLen > longestKey {
			longestKey = keyLen
		}
	}

	formatString := fmt.Sprintf("%s  %%-%ds = %%q\n", whitespace, longestKey)

	sb := new(strings.Builder)
	sb.WriteString("{\n")
	for k, v := range in {
		sb.WriteString(fmt.Sprintf(formatString, k, v))
	}
	sb.WriteString(whitespace + "}")

	return sb.String()
}

func boolPtrOrNull(b *bool) string {
	if b == nil {
		return "null"
	}
	return strconv.FormatBool(*b)
}

func cidrOrNull(in *net.IPNet) string {
	if in == nil {
		return "null"
	}
	return `"` + in.String() + `"`
}

func intPtrOrNull[A constraints.Integer](in *A) string {
	if in == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *in)
}

func stringSetOrNull(in []string) string {
	if len(in) == 0 {
		return "null"
	}

	sb := new(strings.Builder)
	for i, s := range in {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("%q", s))
		} else {
			sb.WriteString(fmt.Sprintf(", %q", s))
		}
	}
	return "[ " + sb.String() + " ]"
}

func randIpNetMust(t testing.TB, cidrBlock string) *net.IPNet {
	t.Helper()

	ip := randIpvAddressMust(t, cidrBlock)

	_, ipNet, _ := net.ParseCIDR(cidrBlock)
	cidrBlockPrefixLen, totalBits := ipNet.Mask.Size()
	targetPrefixLen := rand.Intn((totalBits-1)-cidrBlockPrefixLen) + cidrBlockPrefixLen

	_, result, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", ip.String(), targetPrefixLen))

	return result
}

func randIpvAddressMust(t testing.TB, cidrBlock string) net.IP {
	t.Helper()

	s, err := acctest.RandIpAddress(cidrBlock)
	if err != nil {
		t.Fatal(err)
	}

	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("randIpvAddressMust failed to parse IP address %q", s)
	}

	return ip
}

func randIntSet(t testing.TB, min, max, count int) []int {
	t.Helper()
	require.Greater(t, max, min)

	resultMap := make(map[int]struct{}, count)
	for len(resultMap) < count {
		resultMap[rand.Intn(1+max-min)+min] = struct{}{}
	}

	result := make([]int, count)
	var i int
	for k := range resultMap {
		result[i] = k
		i++
	}

	return result
}

func randomRT(t testing.TB) string {
	t.Helper()

	// three syntactic styles for RTs
	r := rand.Intn(3)
	switch r {
	case 0: // 16-bits:32-bits
		return fmt.Sprintf("%d:%d", uint16(rand.Uint32()), rand.Uint32())
	case 1: // 32-bits:16-bits
		return fmt.Sprintf("%d:%d", rand.Uint32(), uint16(rand.Uint32()))
	case 2: // IPv4:16-bits
		return fmt.Sprintf("%s:%d", randIpvAddressMust(t, "192.0.2.0/24").String(), uint16(rand.Uint32()))
	}

	panic(nil)
}

func randomRTs(t testing.TB, min, max int) []string {
	t.Helper()

	result := make([]string, rand.Intn(max-min)+min)
	for i := range result {
		result[i] = randomRT(t)
	}

	return result
}

func randomIPs(t testing.TB, n int, ipv4Cidr, ipv6Cidr string) []string {
	t.Helper()

	var cidrBlocks []string
	if ipv4Cidr != "" {
		cidrBlocks = append(cidrBlocks, ipv4Cidr)
	}
	if ipv6Cidr != "" {
		cidrBlocks = append(cidrBlocks, ipv6Cidr)
	}

	if len(cidrBlocks) == 0 {
		t.Fatal("cannot make random IPs without any CIDR block")
	}

	result := make([]string, n)
	for i := range result {
		s, err := acctest.RandIpAddress(cidrBlocks[rand.Intn(len(cidrBlocks))])
		require.NoError(t, err)
		result[i] = s
	}

	return result
}

func randomStrings(strCount int, strLen int) []string {
	result := make([]string, strCount)
	for i := 0; i < strCount; i++ {
		result[i] = acctest.RandString(strLen)
	}
	return result
}

func randomJson(t testing.TB, maxInt int, strLen int, count int) json.RawMessage {
	t.Helper()

	preResult := make(map[string]any, count)
	for i := 0; i < count; i++ {
		if rand.Int()%2 == 0 {
			preResult["a"+acctest.RandString(strLen-1)] = rand.Intn(maxInt)
		} else {
			preResult["a"+acctest.RandString(strLen-1)] = acctest.RandString(strLen)
		}
	}

	result, err := json.Marshal(&preResult)
	require.NoError(t, err)

	return result
}

func TestRandomPrefix(t *testing.T) {
	for _ = range 10 {
		//rp := randomPrefix(t, "10.0.0.0/23", 29)
		rp := randomPrefix(t, "2001:db8::/32", 127)
		log.Println(rp.String())
	}
}

func randomPrefix(t testing.TB, cidrBlock string, bits int) net.IPNet {
	t.Helper()

	ip, block, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		t.Fatalf("randomPrefix cannot parse cidrBlock - %s", err)
	}
	if block.IP.String() != ip.String() {
		t.Fatal("invocation of randomPrefix doesn't use a base block address")
	}

	mOnes, mBits := block.Mask.Size()
	if mOnes >= bits {
		t.Fatalf("cannot select a random /%d from within %s", bits, cidrBlock)
	}

	// generate a completely random address
	randomIP := make(net.IP, mBits/8)
	_, err = crand.Read(randomIP)
	if err != nil {
		t.Fatalf("rand read failed")
	}

	// mask off the "network" bits
	for i, b := range randomIP {
		mBitsThisByte := min(mOnes, 8)
		mOnes -= mBitsThisByte

		byteMask := byte(math.MaxUint8 >> mBitsThisByte)

		randomIP[i] = b & byteMask

		block.IP[i] = block.IP[i] | (b & byteMask)
	}

	block.Mask = net.CIDRMask(bits, mBits)

	_, result, err := net.ParseCIDR(block.String())
	if err != nil {
		t.Fatal("failed to parse own CIDR block")
	}

	return *result
}

func randomSlash31(t testing.TB, cidrBlock string) net.IPNet {
	t.Helper()

	ip := randIpvAddressMust(t, cidrBlock)
	_, ipNet, err := net.ParseCIDR(ip.String() + "/31")
	require.NoError(t, err)
	return *ipNet
}

func randomSlash127(t testing.TB, cidrBlock string) net.IPNet {
	t.Helper()

	ip := randIpvAddressMust(t, cidrBlock)
	_, ipNet, err := net.ParseCIDR(ip.String() + "/127")
	require.NoError(t, err)
	return *ipNet
}

type lineNumberer struct {
	lines []string
	base  int
}

func (o *lineNumberer) setBase(base int) error {
	switch base {
	case 2:
	case 8:
	case 10:
	case 16:
	default:
		return fmt.Errorf("base %d not supported", base)
	}

	o.base = base
	return nil
}

func (o *lineNumberer) append(l string) {
	o.lines = append(o.lines, l)
}

func (o *lineNumberer) appendf(format string, a ...any) {
	o.append(fmt.Sprintf(format, a...))
}

func (o *lineNumberer) string() string {
	count := len(o.lines)
	if count == 0 {
		return ""
	}

	base := o.base
	if base == 0 {
		base = 10
	}

	formatStr, _ := padFormatStr(count, base) // err ignored because only valid base can exist here

	sb := new(strings.Builder)
	for i, line := range o.lines {
		sb.WriteString(fmt.Sprintf(formatStr, i+1) + " " + line + "\n")
	}

	return sb.String()
}

func padFormatStr(n, base int) (string, error) {
	var baseChar string
	switch base {
	case 2:
		baseChar = "b"
	case 8:
		baseChar = "o"
	case 10:
		baseChar = "d"
	case 16:
		baseChar = "x"
	default:
		return "", fmt.Errorf("base %d not supported", base)
	}

	c := int(math.Floor(math.Log(float64(n))/math.Log(float64(base)))) + 1
	return fmt.Sprintf("%%%d%s", c, baseChar), nil
}

func TestPadFormatStr(t *testing.T) {
	type testCase struct {
		n        int
		base     int
		expected string
	}

	testCases := []testCase{
		{n: 1, base: 2, expected: "%1b"},
		{n: 2, base: 2, expected: "%2b"},
		{n: 3, base: 2, expected: "%2b"},
		{n: 4, base: 2, expected: "%3b"},
		{n: 7, base: 2, expected: "%3b"},

		{n: 1, base: 8, expected: "%1o"},
		{n: 7, base: 8, expected: "%1o"},
		{n: 8, base: 8, expected: "%2o"},
		{n: 63, base: 8, expected: "%2o"},
		{n: 64, base: 8, expected: "%3o"},
		{n: 511, base: 8, expected: "%3o"},

		{n: 1, base: 10, expected: "%1d"},
		{n: 9, base: 10, expected: "%1d"},
		{n: 10, base: 10, expected: "%2d"},
		{n: 99, base: 10, expected: "%2d"},
		{n: 100, base: 10, expected: "%3d"},
		{n: 999, base: 10, expected: "%3d"},

		{n: 1, base: 16, expected: "%1x"},
		{n: 15, base: 16, expected: "%1x"},
		{n: 16, base: 16, expected: "%2x"},
		{n: 255, base: 16, expected: "%2x"},
		{n: 256, base: 16, expected: "%3x"},
		{n: 4095, base: 16, expected: "%3x"},
	}

	for _, tCase := range testCases {
		tCase := tCase
		t.Run(fmt.Sprintf("%d_with_base_%d", tCase.n, tCase.base), func(t *testing.T) {
			result, err := padFormatStr(tCase.n, tCase.base)
			require.NoError(t, err)
			assert.Equal(t, tCase.expected, result)
		})
	}
}

func TestLineNumbererString(t *testing.T) {
	type testCase struct {
		lines    []string
		base     int
		expected string
	}

	testCases := []testCase{
		{
			lines:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
			base:     2,
			expected: "    1 a\n   10 b\n   11 c\n  100 d\n  101 e\n  110 f\n  111 g\n 1000 h\n 1001 i\n 1010 j\n 1011 k\n 1100 l\n 1101 m\n 1110 n\n 1111 o\n10000 p\n",
		},
		{
			lines:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
			base:     8,
			expected: " 1 a\n 2 b\n 3 c\n 4 d\n 5 e\n 6 f\n 7 g\n10 h\n11 i\n12 j\n13 k\n14 l\n15 m\n16 n\n17 o\n20 p\n",
		},
		{
			lines:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
			base:     10,
			expected: " 1 a\n 2 b\n 3 c\n 4 d\n 5 e\n 6 f\n 7 g\n 8 h\n 9 i\n10 j\n11 k\n12 l\n13 m\n14 n\n15 o\n16 p\n",
		},
		{
			lines:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
			base:     16,
			expected: " 1 a\n 2 b\n 3 c\n 4 d\n 5 e\n 6 f\n 7 g\n 8 h\n 9 i\n a j\n b k\n c l\n d m\n e n\n f o\n10 p\n",
		},
	}

	for _, tCase := range testCases {
		ln := new(lineNumberer)
		require.NoError(t, ln.setBase(tCase.base))
		for _, line := range tCase.lines {
			ln.append(line)
		}
		result := ln.string()

		assert.Equal(t, tCase.expected, result)
	}
}

func newTestChecks(path string) testChecks {
	return testChecks{path: path}
}

type testChecks struct {
	path     string
	logLines lineNumberer
	checks   []resource.TestCheckFunc
}

func (o *testChecks) setPath(path string) {
	o.path = path
}

func (o *testChecks) append(t testing.TB, testCheckFuncName string, testCheckFuncArgs ...string) {
	t.Helper()

	switch testCheckFuncName {
	case "TestCheckResourceAttrSet":
		if len(testCheckFuncArgs) != 1 {
			t.Fatalf("%s requires 1 args, got %d", testCheckFuncName, len(testCheckFuncArgs))
		}
		o.checks = append(o.checks, resource.TestCheckResourceAttrSet(o.path, testCheckFuncArgs[0]))
		o.logLines.appendf("TestCheckResourceAttrSet(%s, %q)", o.path, testCheckFuncArgs[0])
	case "TestCheckNoResourceAttr":
		if len(testCheckFuncArgs) != 1 {
			t.Fatalf("%s requires 1 args, got %d", testCheckFuncName, len(testCheckFuncArgs))
		}
		o.checks = append(o.checks, resource.TestCheckNoResourceAttr(o.path, testCheckFuncArgs[0]))
		o.logLines.appendf("TestCheckNoResourceAttr(%s, %q)", o.path, testCheckFuncArgs[0])
	case "TestCheckResourceAttr":
		if len(testCheckFuncArgs) != 2 {
			t.Fatalf("%s requires 2 args, got %d", testCheckFuncName, len(testCheckFuncArgs))
		}
		o.checks = append(o.checks, resource.TestCheckResourceAttr(o.path, testCheckFuncArgs[0], testCheckFuncArgs[1]))
		o.logLines.appendf("TestCheckResourceAttr(%s, %q, %q)", o.path, testCheckFuncArgs[0], testCheckFuncArgs[1])
	case "TestCheckTypeSetElemAttr":
		if len(testCheckFuncArgs) != 2 {
			t.Fatalf("%s requires 2 args, got %d", testCheckFuncName, len(testCheckFuncArgs))
		}
		o.checks = append(o.checks, resource.TestCheckTypeSetElemAttr(o.path, testCheckFuncArgs[0], testCheckFuncArgs[1]))
		o.logLines.appendf("TestCheckTypeSetElemAttr(%s, %q, %q)", o.path, testCheckFuncArgs[0], testCheckFuncArgs[1])
	case "TestCheckResourceAttrPair":
		if len(testCheckFuncArgs) != 2 {
			t.Fatalf("%s requires 2 args, got %d", testCheckFuncName, len(testCheckFuncArgs))
		}
		o.checks = append(o.checks, resource.TestCheckResourceAttrPair(o.path, testCheckFuncArgs[0], o.path, testCheckFuncArgs[1]))
		o.logLines.appendf("TestCheckResourceAttrPair(%s, %q, %s, %q)", o.path, testCheckFuncArgs[0], o.path, testCheckFuncArgs[1])
	case "TestCheckResourceInt64AttrBetween":
		if len(testCheckFuncArgs) != 3 {
			t.Fatalf("%s requires 3 args, got %d", testCheckFuncName, len(testCheckFuncArgs))
		}

		int64min, err := strconv.ParseInt(testCheckFuncArgs[1], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("TestCheckResourceInt64AttrBetween min value %q does not parse to int64 - %s", testCheckFuncArgs[1], err))
		}
		int64max, err := strconv.ParseInt(testCheckFuncArgs[2], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("TestCheckResourceInt64AttrBetween max value %q does not parse to int64 - %s", testCheckFuncArgs[2], err))
		}

		o.checks = append(o.checks, testcheck.TestCheckResourceInt64AttrBetween(o.path, testCheckFuncArgs[0], int64min, int64max))
		o.logLines.appendf("TestCheckResourceInt64AttrBetween(%s, %q, %d, %d)", o.path, testCheckFuncArgs[0], int64min, int64max)
	}
}

func (o *testChecks) appendSetNestedCheck(_ testing.TB, attrName string, m map[string]string) {
	o.checks = append(o.checks, resource.TestCheckTypeSetElemNestedAttrs(o.path, attrName, m))
	o.logLines.appendf("TestCheckTypeSetElemNestedAttrs(%s, %s, %s)", o.path, attrName, m)
}

func (o *testChecks) extractFromState(t testing.TB, path, id string, targetMap map[string]string) {
	o.checks = append(o.checks, extractValueFromTerraformState(t, path, id, targetMap))
	o.logLines.appendf("extractValueFromTerraformState(%s, %q)", path, id)
}

func (o *testChecks) string() string {
	return o.logLines.string()
}

func extractValueFromTerraformState(t testing.TB, name string, id string, targetMap map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		targetMap[t.Name()] = rs.Primary.Attributes[id]

		return nil
	}
}
