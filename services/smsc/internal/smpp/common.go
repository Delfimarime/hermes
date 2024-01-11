package smpp

const (
	smppClientEventTypeAttribute    = "smpp_client_event_type"
	smppClientEventProblemAttribute = "smpp_client_event_problem"

	smscIdAttribute    = "smsc_id"
	smscNameAttribute  = "smsc_name"
	smscAliasAttribute = "smsc_alias"
	smscStateAttribute = "smsc_state"
	smscTypeAttribute  = "smsc_type"

	smsIdAttribute          = "sms_id"
	smsDestinationAttribute = "sms_message"

	errorAttribute            = "caused-by"
	pduAttribute              = "pdu"
	pduFieldAttribute         = pduAttribute + "_field"
	pduHeaderAttribute        = pduAttribute + "_header"
	pduHeaderIdAttribute      = pduHeaderAttribute + "_id"
	pduHeaderStatusAttribute  = pduHeaderAttribute + "_status"
	pduFieldEsmClassAttribute = pduFieldAttribute + "_esm_class"
)

func truncate(str string, maxSize int) string {
	if len(str) < maxSize {
		return str
	}
	return str[0 : maxSize-1]
}
