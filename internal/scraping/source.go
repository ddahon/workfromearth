package scraping

type AtsKind int16

const (
	UnknownAts AtsKind = iota
	AshbyAts
)

type UnknownSource struct {
	Url string
}

func (source UnknownSource) Company() string {
	return ""
}

func (source UnknownSource) Scrape() ([]Job, error) {
	return []Job{}, nil
}
