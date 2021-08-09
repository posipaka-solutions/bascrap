package scraper

type NoNewsUpdate struct{}

func (err *NoNewsUpdate) Error() string {
	return "News did not changed since last time."
}
