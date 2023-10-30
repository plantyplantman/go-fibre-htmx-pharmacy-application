package presenter

type Table struct {
	Headers []string
	Rows    []*Row
}

type Row struct {
	Cells []*Cell
	Class string
}

type Cell string

func NewCell(s string) *Cell {
	c := Cell(s)
	return &c
}

func (t *Table) DeleteAtIdx(idx int) {
	t.Rows = append(t.Rows[:idx], t.Rows[idx+1:]...)
}

func (t *Table) AddRow(r *Row) {
	t.Rows = append(t.Rows, r)
}
func (t *Table) AddRows(r []*Row) {
	t.Rows = append(t.Rows, r...)
}

func (t *Table) AddHeader(h string) {
	t.Headers = append(t.Headers, h)
}
