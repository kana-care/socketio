package socketio

func amp(s string) *string { return &s }

func stoi(s []string) []interface{} {
	rtn := make([]interface{}, len(s))
	for i, v := range s {
		rtn[i] = v
	}
	return rtn
}

func boolIs(a, b bool) bool { return !(a == b && !a) } // 0 == 0 && 0 then false... otherwise true

func serviceError(err error) []interface{} { return []interface{}{Error(err)} }

func scrub(strOnly bool, event Event, data []Serializable) (out interface{}, cb EventCallback, err error) {
	if strOnly {
		rtn := make([]string, len(data)+1)
		rtn[0] = event
		for i, v := range data {
			if cbv, ok := v.(EventCallback); ok && i == len(data)-1 {
				return rtn[:len(rtn)-1], cbv, nil
			}
			rtn[i+1], err = v.Serialize()
			if err != nil {
				return nil, cb, ErrBadScrub.F(err)
			}
		}
		return rtn, nil, nil
	}
	type ifa interface{ Interface() interface{} }
	rtn := make([]interface{}, len(data)+1)
	rtn[0] = event
	for i, v := range data {
		if cbv, ok := v.(EventCallback); ok && i == len(data)-1 {
			return rtn[:len(rtn)-1], cbv, nil
		}
		if vi, ok := v.(ifa); ok {
			rtn[i+1] = vi.Interface()
			if err, ok := rtn[i+1].(error); ok {
				rtn[i+1] = err.Error()
			}
			continue
		}
		rtn[i+1] = v
	}
	return rtn, nil, nil
}
