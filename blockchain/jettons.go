package blockchain

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

const tonAPI = "https://tonapi.io/v2/jettons/"

type JettonData struct {
	TotalSupply string `json:"total_supply"`
}

func GetTotalSupply() uint64 {
	resp, err := http.Get(tonAPI + "0:42a3dab99606812e24cf919c056757656769791a0efa6d3e7f7939a5d1fcd9c9")
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var data JettonData
	json.Unmarshal(body, &data)

	TotalSupply, _ := strconv.ParseUint(data.TotalSupply, 10, 0)

	return TotalSupply
}
