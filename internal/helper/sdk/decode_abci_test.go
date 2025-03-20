package sdkhelper

// func TestDecodeABCIData(t *testing.T) {
// 	// Test for one message in the tx
// 	dataStr1 := "EjAKLi9iYWJ5bG9uLmZpbmFsaXR5LnYxLk1zZ0FkZEZpbmFsaXR5U2lnUmVzcG9uc2U="
// 	dataBz, err := base64.StdEncoding.DecodeString(dataStr1)
// 	assert.NoError(t, err)
// 	result, err := DecodeABCIData(dataBz)
// 	assert.NoError(t, err)
// 	for _, msg := range result.MsgResponses {
// 		t.Logf("decodedData: %v", msg.TypeUrl)
// 	}

// 	// Test for multiple messages in the tx
// 	dataStr2 := "EiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZRImCiQvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kUmVzcG9uc2USJgokL2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZFJlc3BvbnNlEiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZRImCiQvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kUmVzcG9uc2USJgokL2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZFJlc3BvbnNlEiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZRImCiQvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kUmVzcG9uc2USJgokL2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZFJlc3BvbnNlEiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZRImCiQvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kUmVzcG9uc2USJgokL2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZFJlc3BvbnNlEiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZRImCiQvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kUmVzcG9uc2USJgokL2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZFJlc3BvbnNlEiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZRImCiQvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kUmVzcG9uc2USJgokL2Nvc21vcy5iYW5rLnYxYmV0YTEuTXNnU2VuZFJlc3BvbnNlEiYKJC9jb3Ntb3MuYmFuay52MWJldGExLk1zZ1NlbmRSZXNwb25zZQ=="
// 	dataBz, err = base64.StdEncoding.DecodeString(dataStr2)
// 	assert.NoError(t, err)
// 	result, err = DecodeABCIData(dataBz)
// 	assert.NoError(t, err)
// 	for _, msg := range result.MsgResponses {
// 		t.Logf("decodedData: %v", msg.TypeUrl)
// 	}
// }

// func TestExtractType(t *testing.T) {
// 	samples := []string{
// 		"/babylon.finality.v1.MsgAddFinalitySigResponse",
// 		"/cosmos.bank.v1beta1.MsgSendResponse",
// 	}

// 	assert.Equal(t, "/babylon.finality.v1.MsgAddFinalitySig", ExtractMsgTypeInResponse(samples[0]))
// 	assert.Equal(t, "/cosmos.bank.v1beta1.MsgSend", ExtractMsgTypeInResponse(samples[1]))
// }
