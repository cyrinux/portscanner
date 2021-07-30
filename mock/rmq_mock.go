package mock

// CreateMockRMQ create RMQ mock
func CreateMockRMQ() MockDelivery {
	return MockDelivery{}
}

type MockDelivery struct {
	Acks     int
	Rejected int
}

func (r *MockDelivery) Payload() string {
	return ""
}
func (r *MockDelivery) Ack() error {
	r.Acks += 1
	return nil
}
func (r *MockDelivery) Reject() error {
	r.Rejected += 1
	return nil
}
func (r *MockDelivery) Push() error {
	return nil
}
