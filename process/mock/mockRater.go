package mock

type mockRater struct {
}

func (msg *mockRater) UpdateRating() error {
	return nil
}
