package cmd

import (
	"github.com/cretz/go-safeclient/client"
	"github.com/olekukonko/tablewriter"
	"io"
	"time"
	"strconv"
	"sort"
)

func writeDirResponseTable(writer io.Writer, dir client.DirResponse) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Name", "Size", "Created On", "Modified On"})
	table.SetBorder(false)
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetAutoFormatHeaders(false)
	table.Append([]string{"./", "",
		dir.Info.CreatedOn.Time().Format(time.RFC822), dir.Info.ModifiedOn.Time().Format(time.RFC822)})
	// Sort the things first
	sort.Sort(dir.SubDirs)
	for _, sub := range dir.SubDirs {
		table.Append([]string{sub.Name + "/", "",
			sub.CreatedOn.Time().Format(time.RFC822), sub.ModifiedOn.Time().Format(time.RFC822)})
	}
	sort.Sort(dir.Files)
	for _, file := range dir.Files {
		table.Append([]string{file.Name, strconv.FormatInt(file.Size, 10),
			file.CreatedOn.Time().Format(time.RFC822), file.ModifiedOn.Time().Format(time.RFC822)})
	}
	table.Render()
}