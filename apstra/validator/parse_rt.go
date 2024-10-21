package apstravalidator

import (
	"context"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	rtSep       = ":"
	rtFormatErr = `A Route Target must take one of the following forms (leading zeros not permitted):
 - <2-byte-value>:<4-byte-value>
 - <4-byte-value>:<2-byte-value>
 - <IPv4-address>:<2-byte-value>
`
)

var _ validator.String = ParseRtValidator{}

type ParseRtValidator struct{}

func (o ParseRtValidator) Description(_ context.Context) string {
	return "Ensures that the supplied can be parsed as a Route Target"
}

func (o ParseRtValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseRtValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// split the RT string
	parts := strings.Split(req.ConfigValue.ValueString(), rtSep)
	if len(parts) != 2 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, rtFormatErr, req.ConfigValue.ValueString()))
		return
	}

	var ipFound bool
	// check to see if part 1 is an IPv4 address
	ip := net.ParseIP(parts[0])
	if ip != nil {
		ipFound = true // we got an IP!

		// is it IPv4?
		if len(ip.To4()) != net.IPv4len {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path, rtFormatErr, req.ConfigValue.ValueString()))
			return
		}
	}

	// make sure "part 1" has length and no prepended zeros (if we haven't already decided it's an IPv4 address)
	if !ipFound && (len(parts[0]) == 0 || (len(parts[0]) >= 2 && strings.HasPrefix(parts[0], "0"))) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, rtFormatErr, req.ConfigValue.ValueString()))
		return
	}

	// make sure "part 2" has length and no prepended zeros
	if len(parts[1]) == 0 || (len(parts[1]) >= 2 && strings.HasPrefix(parts[1], "0")) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, rtFormatErr, req.ConfigValue.ValueString()))
		return
	}

	// determine whether we've got a 32-bit "first part", and thus require a 16-bit "second part"
	var firstPartIs32bits bool
	if ipFound {
		firstPartIs32bits = true
	} else {
		// try parsing p1 as a 32-bit value
		p1, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path, rtFormatErr, req.ConfigValue.ValueString()))
			return
		}

		// does p1 require 32 bits?
		if p1 > math.MaxUint16 {
			firstPartIs32bits = true
		}
	}

	// try parsing p2 as a 32-bit value
	p2, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, rtFormatErr, req.ConfigValue.ValueString()))
		return
	}

	// p1 and p2 can't both be 32 bits
	if firstPartIs32bits && p2 > math.MaxUint16 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, rtFormatErr, req.ConfigValue.ValueString()))
		return
	}
}

func ParseRT() validator.String {
	return ParseRtValidator{}
}
