package http

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/processor/provision/pipeline"
	"html/template"
	"net/http"
)

func buildStatisticsPage(w http.ResponseWriter, statistics []*pipeline.Summary) {

	tmpl := `<html>
		<head>
		<title>/debug/statistics</title>
		<style>
		.profile-name{
			display:inline-block;
			width:6rem;
		}
		table, th, td {
		  border: 1px solid black;
		  border-collapse: collapse;
		}
		th, td {
		  padding: 5px;
		}
		</style>
		<meta http-equiv="refresh" content="1" />
		</head>
		<body>
		/debug/statistics/
		<br>
		<p>Set debug=1 as a query parameter to export in legacy text format</p>
		<br>
		<br>
		<table>
			<tr>
				<th>Namespace</th>
				<th>Function</th>
				<th>Run</th>
				<th>E</th>
				<th>State</th>
				<th>ET</th>
				<th>Processed</th>
				<th>Invalid</th>
				<th>T</th>
				<th>State</th>
				<th>TL</th>
				<th>Processed</th>
				<th>Invalid</th>
				<th>L</th>
			</tr>
			{{range .Items}}
			<tr>
				<td>{{ .Namespace }}</td>
				<td>{{ .Function }}</td>
				<td>{{ .Run }}</td>
				<td>{{ .Statistics.Threads.NumProvisionedExtractRoutines }}</td>
				<td>{{ .ETState }}</td>
				<td>{{ .ETSize }}</td>
				<td>{{ .Statistics.Data.TotalOverETChannel }}</td>
				<td>{{ .Statistics.Data.TotalInvalidOverETChannel }}</td>
				<td>{{ .Statistics.Threads.NumActiveTransformRoutines }}</td>
				<td>{{ .TLState }}</td>
				<td>{{ .TLSize }}</td>
				<td>{{ .Statistics.Data.TotalOverTLChannel }}</td>
				<td>{{ .Statistics.Data.TotalInvalidOverTLChannel }}</td>
				<td>{{ .Statistics.Threads.NumActiveLoadRoutines }}</td>
			</tr>
			{{end}}
		</table>
		<br>
		<br>
		<p>Profile Description:</p>
		<ul>
		  <li><b>Run</b>&emsp;The instance of the running cluster.</li>
		  <li><b>E</b>&emsp;The number of extract functions loaded.</li>
          <li><b>T</b>&emsp;The number of transform functions loaded.</li>
          <li><b>L</b>&emsp;The number of extract functions loaded.</li>
		  <li><b>ET</b>&emsp;The number of records being sent from E to T functions.</li>
          <li><b>TL</b>&emsp;The number of records being sent from T to L functions.</li>
		  <li><b>State</b>&emsp;How the runner will behave to the incoming data; when congested the number of functions will shrink, shrink when idle.</li>
		  <li><b>Processed</b>&emsp;The number of records sent over a channel.</li>
		  <li><b>Invalid</b>&emsp;The number of records dropped on a channel; This should never be > 1.</li>
		</ul>
		</body>
		</html>
	`
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Items []*pipeline.Summary
	}{
		Title: "Statistics",
		Items: statistics,
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
