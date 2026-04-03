package output

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

func PrintJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

func PrintTable(w io.Writer, headers []string, rows [][]string) {
	table := tablewriter.NewWriter(w)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("  ")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(rows)
	table.Render()
}

func FormatTime(unixTs int64) string {
	if unixTs == 0 {
		return "-"
	}
	return time.Unix(unixTs, 0).Format("2006-01-02 15:04:05")
}

func FormatQuota(unlimited bool, quota int64) string {
	if unlimited {
		return "unlimited"
	}
	return strconv.FormatInt(quota, 10)
}