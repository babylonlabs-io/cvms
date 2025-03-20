package sdkhelper

// import (
// 	"regexp"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/gogoproto/proto"
// )

/*
NOTE: this function is decode the data field of ABCI Check/DeliverTx response.

	ref; https://github.com/cosmos/cosmos-sdk/blob/main/baseapp/baseapp.go#L1137
	// makeABCIData generates the Data field to be sent to ABCI Check/DeliverTx.
	func makeABCIData(msgResponses []*codectypes.Any) ([]byte, error) {
		return proto.Marshal(&sdk.TxMsgData{MsgResponses: msgResponses})
	}
*/
// func DecodeABCIData(bz []byte) (sdk.TxMsgData, error) {
// 	var txMsgData sdk.TxMsgData
// 	if err := proto.Unmarshal(bz, &txMsgData); err != nil {
// 		return sdk.TxMsgData{}, err
// 	}

// 	return txMsgData, nil
// }

// func ExtractMsgTypeInResponse(msg string) string {
// 	re := regexp.MustCompile(`^(\/[\w.]+)Response$`)
// 	return re.ReplaceAllString(msg, `$1`)
// }
