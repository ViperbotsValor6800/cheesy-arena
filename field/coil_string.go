// Code generated by "stringer -type=coil"; DO NOT EDIT.

package field

import "strconv"

const _coil_name = "heartbeatmatchResetstackLightGreenstackLightOrangestackLightRedstackLightBluered1EthernetDisablered2EthernetDisablered3EthernetDisableblue1EthernetDisableblue2EthernetDisableblue3EthernetDisablecoilCount"

var _coil_index = [...]uint8{0, 9, 19, 34, 50, 63, 77, 96, 115, 134, 154, 174, 194, 203}

func (i coil) String() string {
	if i < 0 || i >= coil(len(_coil_index)-1) {
		return "coil(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _coil_name[_coil_index[i]:_coil_index[i+1]]
}
