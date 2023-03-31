package tfapstra

//func TestSliceAttrValueToSliceString(t *testing.T) {
//	test := []attr.Value{
//		types.String{Value: "foo"},
//		types.Int64{Value: 6},
//	}
//	expected := []string{
//		"foo",
//		"6",
//	}
//	result := sliceAttrValueToSliceString(test)
//	if len(expected) != len(result) {
//		t.Fatalf("expected %d results, got %d results", len(expected), len(result))
//	}
//	for i := 0; i < len(expected); i++ {
//		if expected[i] != result[i] {
//			t.Fatalf("expected '%s', got '%s'", expected[i], result[i])
//		}
//	}
//}

//func TestSliceAttrValueToSliceObjectId(t *testing.T) {
//	test := []attr.Value{
//		types.String{Value: "foo"},
//		types.Int64{Value: 6},
//	}
//	expected := []apstra.ObjectId{
//		"foo",
//		"6",
//	}
//	result := sliceAttrValueToSliceObjectId(test)
//	if len(expected) != len(result) {
//		t.Fatalf("expected %d results, got %d results", len(expected), len(result))
//	}
//	for i := 0; i < len(expected); i++ {
//		if expected[i] != result[i] {
//			t.Fatalf("expected '%s', got '%s'", expected[i], result[i])
//		}
//	}
//}

//func TestSliceWithoutString(t *testing.T) {
//	type testDatum struct {
//		test     []string
//		target   string
//		expected []string
//		n        int
//	}
//	var testData []testDatum
//	testData = append(testData, testDatum{[]string{"foo", "bar", "baz"}, "bogus", []string{"foo", "bar", "baz"}, 0})
//	testData = append(testData, testDatum{[]string{"foo", "bar", "baz"}, "bar", []string{"foo", "baz"}, 1})
//	testData = append(testData, testDatum{[]string{"foo", "bar", "bar", "baz"}, "bar", []string{"foo", "baz"}, 2})
//	testData = append(testData, testDatum{[]string{"bar", "foo", "bar", "bar", "baz", "bar"}, "bar", []string{"foo", "baz"}, 4})
//	testData = append(testData, testDatum{[]string{"bar", "bar", "bar", "bar", "bar", "bar"}, "bar", []string{}, 6})
//	testData = append(testData, testDatum{[]string{"foo", "bazbarbaz", "baz"}, "bar", []string{"foo", "bazbarbaz", "baz"}, 0})
//
//	for i, td := range testData {
//		result, n := sliceWithoutString(td.test, td.target)
//		if len(td.expected) != len(result) {
//			t.Fatalf("testData[%d] expected '%s', got '%s'", i, strings.Join(td.expected, ","), strings.Join(result, ","))
//		}
//		for i := range result {
//			if td.expected[i] != result[i] {
//				t.Fatalf("testData[%d] expected '%s', got '%s'", i, strings.Join(td.expected, ","), strings.Join(result, ","))
//			}
//		}
//		if td.n != n {
//			t.Fatalf("testData[%d] expected '%d', got '%d'", i, td.n, n)
//		}
//	}
//}
