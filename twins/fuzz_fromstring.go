package twins

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func getTypeString(str string) (string, string) {
	str = strings.Trim(str, "\t\n ")
	i := strings.IndexAny(str, "{\n")
	return str[:i], str[i:]
}

func getFieldString(str string) (string, string) {
	str = strings.Trim(str, "\t\n ")
	fmt.Println(str)
	i := strings.IndexAny(str, ":")
	fmt.Println(str, i)
	return str[:i], str[i:]
}

func getValue(str string) (string, string) {
	str = strings.Trim(str, "\t\n ")
	i := strings.IndexAny(str, "\n")
	return str[:i], str[i:]
}

func VoteMsgFromString(str string) (voteMsg VoteMsg) {
	for {
		i := strings.IndexAny(str, "{}\n")
		str = str[i:]
		if str[0] == '}' {
			return voteMsg
		}

		var fieldStr string
		fieldStr, str = getFieldString(str)
		// pointer to struct - addressable
		ps := reflect.ValueOf(&voteMsg)
		// struct
		s := ps.Elem()
		f := s.FieldByName(fieldStr)
		if f.IsValid() {
			if f.CanSet() {
				if f.Kind() == reflect.Int {

					var valueStr string
					valueStr, str = getValue(str)

					fmt.Println(valueStr)

					i, err := strconv.Atoi(valueStr)

					if err != nil {
						return VoteMsg{}
					}

					x := int64(i)
					if !f.OverflowInt(x) {
						f.SetInt(x)
					}
				}
			}
		}

	}
}

func fuzzMsgFromString(str string) FuzzMsg {
	typeStr, restStr := getTypeString(str)
	var retMsg FuzzMsg

	switch typeStr {
	case "twins.VoteMsg":
		msg := VoteMsgFromString(restStr)
		retMsg = &msg
	case "twins.ProposeMsg":
		msg := ProposeMsg{}
		retMsg = &msg
	case "twins.NewViewMsg":
		msg := NewViewMsg{}
		retMsg = &msg
	case "twins.TimeoutMsg":
		msg := TimeoutMsg{}
		retMsg = &msg
	default:

	}

	return retMsg
}
