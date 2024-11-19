package baiduyun

type Response struct {
	TotalCount int                      `json:"totalCount"`
	Result     []map[string]interface{} `json:"result"`
}
