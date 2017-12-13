package controllers

import (
	"log"
	"strconv"
	"router"
	"net/http"
	"strings"
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"encoding/json"
)

// Load forces the program to call all the init() funcs in each models file
func Load(schema *graphql.Schema) {

	if schema != nil {
		router.Get("/", GraphIql)
		router.PostHandler("/query", &relay.Handler{Schema: schema})
	} else {
		router.Get("/", Welcome)
	}
}

func Welcome(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Welcome")
}

func GraphIql(w http.ResponseWriter, req *http.Request) {
	w.Write(page)
}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(value) {
			return true
		}
	}
	return false
}

func StringToUInt(ID string) uint {
	u64, err := strconv.ParseUint(ID, 10, 32)
	if err != nil {
		log.Println(err)
		return 0
	}
	wd := uint(u64)
	return wd
}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/query", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
