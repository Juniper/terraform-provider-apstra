package importvalidation

type IdentityWithAttributes = identityWithAttributes

const (
	FailSummaryGeneral      = failSummaryGeneral
	FailSummaryMissingField = failSummaryMissingField
	FailSummaryInvalidField = failSummaryInvalidField
	FailSummaryInvalidType  = failSummaryInvalidType
	FailSummaryMissingValue = failSummaryMissingValue

	FailDetailNilIdentity   = failDetailNilIdentity
	FailDetailNotStruct     = failDetailNotStruct
	FailDetailNoTag         = failDetailNoTag
	FailDetailCannotInspect = failDetailCannotInspect
	FailDetailNotAttrValue  = failDetailNotAttrValue
	FailDetailNullOrUnknown = failDetailNullOrUnknown
)
