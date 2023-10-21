package apio

type OpenApi3 struct {
	Openapi    string                 `json:"openapi"`
	Info       map[string]interface{} `json:"info"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

func ToOpenApi3(api Api) string {

	//openApi := OpenApi3{
	//	Openapi: "3.0.0",
	//	Info: map[string]interface{}{
	//		"title":       api.Name,
	//		"description": api.Description,
	//		"version":     api.Version,
	//	},
	//	Paths:      api.GetPaths(),
	//	Components: api.GetComponents(),
	//}
	//
	//jsonString, _ := json.Marshal(openApi)

	//return string(jsonString)

	return "" // TODO: Implement
}
