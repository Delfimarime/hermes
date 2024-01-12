package smpp

const (
	smppClientEventTypeAttribute    = "smpp_client_event_type"
	smppClientEventProblemAttribute = "smpp_client_event_problem"

	SmscIdAttribute    = "smsc_id"
	SmscNameAttribute  = "smsc_name"
	SmscAliasAttribute = "smsc_alias"
	SmscTypeAttribute  = "smsc_type"
	smscStateAttribute = "smsc_state"

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
