package mock

// type Delivery interface {
// 	Payload() string

// 	Ack() error
// 	Reject() error
// 	Push() error
// }

// CreateMockRMQ create RMQ mock
func CreateMockRMQ() MockDelivery {
	return MockDelivery{}
}

type MockDelivery struct {
}

func (r *MockDelivery) Payload() string {
	return ""
}
func (r *MockDelivery) Ack() error {
	return nil
}
func (r *MockDelivery) Reject() error {
	return nil
}
func (r *MockDelivery) Push() error {
	return nil
}
