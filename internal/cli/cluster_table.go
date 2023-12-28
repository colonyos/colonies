package cli

import (
	"strconv"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/muesli/termenv"
)

func printClusterInfoTable(cluster *cluster.Config) {
	t, theme := createTable(1)

	var cols = []table.Column{
		{ID: "name", Name: "Name", SortIndex: 1},
		{ID: "host", Name: "Host", SortIndex: 2},
		{ID: "apiport", Name: "APIPort", SortIndex: 3},
		{ID: "etcdclientport", Name: "EtcdClientPort", SortIndex: 4},
		{ID: "etcdpeerport", Name: "EtcdPeerPort", SortIndex: 5},
		{ID: "relayport", Name: "RelayPort", SortIndex: 6},
		{ID: "leader", Name: "Leader", SortIndex: 6},
	}
	t.SetCols(cols)

	for _, node := range cluster.Nodes {
		row := []interface{}{
			termenv.String(node.Name).Foreground(theme.ColorCyan),
			termenv.String(node.Host).Foreground(theme.ColorViolet),
			termenv.String(strconv.Itoa(node.APIPort)).Foreground(theme.ColorBlue),
			termenv.String(strconv.Itoa(node.EtcdClientPort)).Foreground(theme.ColorMagenta),
			termenv.String(strconv.Itoa(node.EtcdPeerPort)).Foreground(theme.ColorYellow),
			termenv.String(strconv.Itoa(node.RelayPort)).Foreground(theme.ColorRed),
			termenv.String(isLeader(cluster.Leader.Name, node.Name)).Foreground(theme.ColorGreen),
		}
		t.AddRow(row)
	}

	t.Render()
}
