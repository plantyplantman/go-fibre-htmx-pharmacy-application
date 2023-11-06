package presenter

type Table struct {
	Headers []string
	Rows    []Row
}

type Row struct {
	Cells     []string
	ClassName string
}

func (t *Table) AppendRows(r ...Row) {
	t.Rows = append(t.Rows, r...)
}

func (t *Table) RowAt(i int) Row {
	return t.Rows[i]
}
