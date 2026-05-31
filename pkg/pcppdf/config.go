package pcppdf

type Config struct {
	TemplateHTML     string
	FontFileName     string
	MetaCSVFile      string
	PCPTabName       string
	ListColumnData   []int
	IndexColumnCod   int
	IndexColItemDesc int
	IndexStartLine   int
	IndexLineDate    int
	FolderOPPCP      string
	FolderOPInclusao string
}

func (c Config) Validate() error {
	if c.TemplateHTML == "" {
		return ErrMissingTemplateHTML
	}
	if c.PCPTabName == "" {
		return ErrMissingPCPTabName
	}
	if len(c.ListColumnData) == 0 {
		return ErrMissingDateColumns
	}
	if c.FolderOPPCP == "" {
		return ErrMissingOutputFolder
	}
	if c.MetaCSVFile == "" {
		return ErrMissingMetadataCSV
	}
	return nil
}
