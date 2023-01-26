package arguments

import (
	"encoding/json"
	"errors"
	"fmt"

	messages "github.com/cucumber/messages/go/v21"
	"sigs.k8s.io/yaml"
)

var (
	ErrTableMustBeWidthTwo = errors.New("DataTable must only have two columns.")
)

type DataTable struct {
	*messages.DataTable
}

func (d *DataTable) MarshalJSON() ([]byte, error) {
	j := map[string]interface{}{}
	for _, row := range d.Rows {
		if len(row.Cells) != 2 {
			return []byte{}, ErrTableMustBeWidthTwo
		}
		key := row.Cells[0].Value
		val := row.Cells[1].Value

		// Get the right type
		var v interface{}
		err := yaml.Unmarshal([]byte(val), &v)
		if err != nil {
			return []byte{}, err
		}

		j[key] = v
	}
	return json.Marshal(j)
}

func (d *DataTable) UnmarshalJSON(b []byte) error {
	j := map[string]interface{}{}
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	i := 0
	d.DataTable = &messages.DataTable{}
	for key, val := range j {
		row := &messages.TableRow{
			Cells: []*messages.TableCell{{Value: key}, {Value: fmt.Sprint(val)}},
		}
		d.DataTable.Rows = append(d.DataTable.Rows, row)
		i++
	}
	return nil
}

func (d *DataTable) UnmarshalInto(o interface{}) error {
	j, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, o)
}
