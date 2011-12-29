package seven5

import (
	"mongrel2"
	"strings"
	"encoding/base64"
	"io"
	"fmt"
//	"os"
	"bytes"
)
type FaviconGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
}

var ico []byte

func init() {
	SetFavicon(defaultIcon)
}

func SetFavicon(b64encodedicon string) {
	ico=decode(b64encodedicon)
}

func decode(encoded string) []byte {
	nows:=strings.Replace(encoded,"\n","",-1/*NO LIMIT*/)
	nows=strings.Replace(nows,"\t","",-1/*NO LIMIT*/)
	r:=strings.NewReader(nows)
	b:=base64.NewDecoder(base64.StdEncoding,r)
	result:=make([]byte,base64.StdEncoding.DecodedLen(len(nows)))
	n,err:=io.ReadFull(b, result)
	if err!=nil && err!=io.ErrUnexpectedEOF{
		panic(fmt.Sprintf("unable to decode the base64 for favico:%s",err.Error()))
	}
	return result[0:n]
}


func (self *FaviconGuise) Name() string {
	return "FaviconGuise" //used to generate the UniqueId so don't change this
}

func (self *FaviconGuise) IsJson() bool {
	return false
}

func (self *FaviconGuise) Pattern() string {
	return "/(favicon.ico)"
}

func (self *FaviconGuise) AppStarting(config *ProjectConfig) error {
	return nil
}

//create a new one... but only one should be needed in any program
func NewFaviconGuise() *FaviconGuise {
	return &FaviconGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new (mongrel2.RawHandlerDefault)}}}
}

//Handle a single request of the HTTP level of mongrel.  This responds with the ico
//
func (self *FaviconGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	path:=req.Header["PATH"]
	path=path[len("/css/"):]
	
	resp:=new (mongrel2.HttpResponse)
	resp.ServerId= req.ServerId
	resp.ClientId = []int{req.ClientId}
	
	resp.ContentLength=len(ico)
	resp.Body = bytes.NewBuffer(ico)
	return resp
}


/* http://nerdycats.wordpress.com/2011/02/28/the-night-i-almost-died-in-paris/*/
//puking by the seine...ahh...
const defaultIcon =`
	AAABAAEAMDAAAAEAIABoJgAAFgAAACgAAAAwAAAAYAAAAAEAIAAAAAAAAAAAAMQO
	AADEDgAAAAAAAAAAAAB8s93/frnf/4rM6f+Kzun/frrg/4C84P+Dw+T/is7p/4TE
	5P9/vOD/iMnn/5DW7/+N0ez/icvo/4nL6P+HyOb/i8/q/47T7f+M0Or/gsHj/4rN
	6f+N0ev/iczo/4/U7P+S2fD/kNnw/5La8v+T3fH/kt3v/5Xd8v+O3e3/Ttu4/0fa
	tP9z3Nf/cdzW/1fbwf9J27T/Otqp/yLanP9L2rT/kN3u/5Xd8f+S3e//kt3w/5Ld
	8P+T3vD/icvp/4TB5f93q9n/fbbd/4nK5/+Kzen/gL3g/4TD4/+GyOX/jtPs/5DV
	7f+Iyef/jNDq/5DX7/+P1u//j9Tt/5DV7f+Hxub/icrp/5La8f+T3PD/jM/r/4vO
	6v+S2vH/kdvz/5Pc9P+T2/D/kt3w/5Pd8f+S3fD/kt3v/5Pd8P+R3e//h93n/43d
	7f+d3fv/iNzp/2rb0P+N3ez/ht3l/2bbyv993N7/lN3x/5Pd8P+S3fD/kt3w/5Lc
	7/+T3/D/jM/q/4nM6f95r9v/erLc/4jL6P+GyOb/iMnn/5HX7v+O0uv/iMnn/5DV
	7v+O0uz/j9Tt/5Pa8P+U3fL/jtLs/43R7P+R2PD/jdLr/5DX7/+T3fD/k9zx/5Lb
	8P+U3vT/jc/j/4O2xf+Pz+D/lODz/5Lc7/+T3fD/kt3w/5Ld7/+T3fD/ld3x/5bd
	8v+A3OH/Vdq//3Hc1f+V3fP/lt3z/5fd9f+W3fT/kt3w/5Ld8P+T3fD/kt3v/5Pd
	8f+U3fL/jdDs/4nL6f96stv/f7rg/4/V7v+P1O3/is7p/43S7f+KzOr/hMPk/43S
	7P+R1+//j9Tu/47S7f+S2O//kNXu/47V7v+T3vP/kNfu/5DV7f+U3/T/ktzv/5Ld
	8P+S2+7/hLbH/4vJ3P+U4PT/ktzv/5Ld8P+S3fD/kt3w/5Pd8P+S3e//kt3v/5nd
	9/9t29D/QNqw/3Lc2v+M3ez/ld3y/5Ld7/+R3e7/kt3w/5Ld8P+S3fD/k93w/5Lb
	8P+N0ez/hMLl/4bG5v98uOL/icrn/4vO6v+P1O3/j9Tt/43S7P+P1O7/is3p/4/U
	7f+S2/H/ktnw/5DW7/+Q1u//kdfv/5DW7/+T3PD/ktzx/5Pe9f+O0eT/gbTD/47W
	6f+T3fD/jMbY/5LY7P+T4PP/ktzu/5Pd8P+S3fD/k93w/5Ld7/+U3fH/j93t/4Lc
	5P9f28X/Tdq4/1vZxf+C2+P/i93q/43d6/+U3fH/kt3w/5Ld8P+S3O//lN/x/43R
	7P+BveL/hMLk/4nL6f+MwM//leX5/4zS7v+Ky+b/ktjv/5LY8P+R1+//kdfu/5LY
	7/+R1+//kdfv/5Lc8v+T3fP/kdnw/5HX7v+R1+//ktrx/5Pe8/+Fvc3/jc7g/5Xi
	9v+R2u7/isTW/5DU6P+U4PP/ktzv/5Ld8P+S3fD/k93w/5Hd7v+V3fL/jtzs/13b
	xP8/2q//WNu+/3vc3f9+3OD/XdvD/4vc6f+V3fL/kt3v/5Ld8P+S3O//lN/z/4/U
	7v+DvuP/iMfo/4vN6/9/Nyf/h5mi/5DW7P+S3ff/ktfu/5HX7/+S2fH/k9rw/5LY
	8P+R1u//kNbv/5HX8P+R1/D/ktnw/5Ha8f+R1/D/ktny/5Pd8v+IxNT/kNfp/5Tg
	8/+S2+7/h8XY/47U5/+U3/L/ktzu/5Pd8P+S3fD/k93w/5Ld7/+U3fH/k9zz/23b
	0f802qj/YtzG/47d7f9M2rb/XtvD/5bd8/+T3fD/ktzw/5Pd8/+S2/L/kNXu/4jK
	5/+Cv+L/iMjo/4zN7P+DKRX/gAAA/4FbWf+OuMf/k+H4/5Hf9v+P1Oz/kdfu/5LZ
	7/+Q2O//ktnw/4/T7f+Lzur/ktnw/5Pe8f+S3PD/k9zw/5Tf9P+GxNb/jNHi/5Xh
	9P+S3O//hsfZ/47V6P+W5Pj/kt3v/5Lc7/+T3fD/kt3w/5Ld8P+S3e//l931/37b
	3v8r2aH/YtzH/2/c0v842qj/gdzg/5fd9P+S3e//ktzw/47V7v+N0Oz/hcTl/4bF
	5v+Hx+f/iMjo/4O/4/+CMSH/gjUo/4IXAP9+FQD/hICF/5LP4v+R3fb/iszo/5DU
	7P+T2/L/ktnw/47T7f+M0Ov/ktrv/5Pd8P+T3fD/kt3v/5Tf9f+Gxtn/i9Hj/5Xh
	9P+R2+7/h8rd/47S5f+Mytr/j9bo/5Tg8/+S3O//k93w/5Pd8P+R3e7/l93z/4Xc
	5f892qv/S9u2/0Lasf9g28X/lN3y/5Pd8P+S3e//k93w/47S7f+Mz+z/jM7q/4vN
	6/+O0u3/iszq/3223/+CLh3/gi4c/4I0Jv+DMSH/gQAA/39EO/+Ioq//jNDr/5Hd
	9v+R2e//jNDq/5HX7v+T2vD/iszp/4jJ6P+S2/D/k97w/5Te9P+Gxdn/jNHj/5Xg
	8/+Q2+//jszf/3u8y/9Zrbn/idHj/5fh9P+R3O7/k93w/5Pd8P+R3e//ld3y/47c
	7P9U27z/Jtqe/0Dar/+F3OX/l930/5Hd7/+S3fD/lN7x/4/V7v+Kzev/j9Tu/43R
	7P+P0+3/jtLt/3644P+CLh3/gi4d/4IuHP+CLx7/gjYp/4IlCP9+AAD/gGpr/47C
	0v+T4fj/j9fv/4/V7P+Q1u7/jtLs/4nL5/+Q1+//kt3x/5Te9P+Gxdj/i9Dh/5Lg
	8/+U4vf/hLPC/0e9yv9E2Or/is/h/5bf8v+R3O//k93w/5Pd8P+S3e//k93w/5Xd
	8v9n28v/Hdqa/2Tbyv+U3fL/lN3w/5Ld7/+T3fD/k93x/4/V7/+Lzuv/jdDs/43R
	7P+P0+7/icvq/3+64f+CLh3/gi4d/4IuHf+CLh3/gi4c/4IyI/+DNCb/gQ4A/3wm
	C/+Fjpb/k9bp/5Dc9v+IyOX/jdHp/43R6/+JzOj/kNjt/5Xg9v+GyNz/jdfp/5vk
	+P+Nxtf/U7vJ/yjt//9G1+n/hsrd/5bf8/+R3fD/k93w/5Pd8P+S3fD/ktzv/5fc
	9P9t28//Mtqm/4Dc4P+Y3fT/kdzt/5Lc7/+T3fH/ktrx/5HY7/+Q1+//j9Tt/47T
	7f+N0e3/iMjo/4K/4/+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gjUo/4Mt
	G/9/AAD/gVFN/4yxv/+T3/X/kt72/5DW7f+O1e3/kdzw/5be8/+Ju8z/iLzM/3e4
	xf9Ov87/OOb7/zfz//8+1uf/gczc/5jg9P+T3O//k9zv/5Lc7v+R3O//kd/x/5ni
	+/94393/U9u5/4rd6f+U3vb/k+L1/5Tf8v+T3fD/ktvx/43R7P+Q1e7/ktjw/47S
	7f+N0e3/isvq/4TC5f+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4c/4Iw
	IP+DNin/gh8A/30UAP+EeX7/kcvd/5Xm+/+S3/H/lt7x/3vD0/9NtsL/QcfX/zPg
	8v807///POr//znv//8z2+z/fMna/4zB0v9+sb3/lODy/5bk+P+Z4fX/k9fp/4/A
	1P9/xsj/ZuDT/4zF0P+Zsbz/icDO/47U6f+U3vT/kNbv/4rN6f+O0uz/kdnv/5DW
	7/+O0u3/h8fo/4C74f+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLhz/gjMl/4MxI/+AAAD/fzYo/4ecpf+Q2u7/meT5/3TC0/8v5fj/NvP//z3t
	//876v//Our+/zvr//816///UMzb/1p7gv9pgYf/i8TT/4zG1f90ucj/V6m3/8LE
	xf/E1NH/VrGy/8rX2f/UzMv/u7Wz/5Wxuf+M1e//kNTv/5HX7v+Q1e3/kdjw/5HZ
	8P+O0uz/hcPl/3+44P+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4c/4IvHv+DNin/gigS/4AAAP+AX1//kcDP/4PM3v8/2u3/N+z//zzq
	//866v//Ouv//zrp/v877///NeP0/zjG0v9B1OX/QtDi/0DM3P813/L/AN/1/6nc
	4//q4+P/x8vO//z7+v/Mzc3/qqip/2Zvff+Gxtz/ktnz/43Q6v+S1/D/kdnx/5LZ
	8P+N0ev/hMLk/4G74/+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLRz/gjEi/4M1KP+BEwD/gi4c/4BzeP9Iy9v/Ne///zvq
	//866v//Ouv//zrp/v867///PN7w/zzU4/857///OPH//zjv//9E8P//Her//2vR
	3//x7u3///////n6+v/e397/3dzc/2tyh/93t8v/lNr0/4K+4v+LzOn/k9vx/5LZ
	8f+Q1e//jNDr/3634P+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHP+CNSj/gzEh/4AAAP9Ql6H/NPD//zvr
	//866v7/Ouv//zrp/v857///PNzv/z7Q4P867P//Ou3//znp/v9C6///Iez//1zY
	6f/n6uv///////v8/P///////////6mqq/9vscX/k9n0/4nJ5/+O0ez/jM7p/4nL
	6P+N0Ov/j9Ps/4O/5P+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gTUo/4YXAP9lYmT/NeP3/zru
	//876f7/Ouv//zrq/v857v//O+T3/z3C0P883e//POb5/zrl+P9C7P//IOv//13X
	5//o6ur///////r8/P//////2NjZ/11nef+AvdP/kNXz/4zP6f+R1u//hMTm/3+8
	3/+Gw+T/hsPl/4XF5v+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gTEi/4UpEf92Kxv/PsTU/zjy
	//876f7/Our//zrr//866v//Ou7//zvX5/8/vMv/QLzL/zzR4P9C7f//G+r//2rT
	4//u7ez////////////+/vv/bG6G/zVeg/+R0un/jNDs/5DW7f+R1u//j9Pt/4vN
	6f+DvuP/fbXf/4G84f+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gjAf/4IzJP+CAAD/VYmR/zTu
	//876///Our//zrr//866///Our//zru//867f//Ouz//znr//9B7P//Fen//3fV
	4//18fD///////////+prLL/Gixj/32xxf+R1fL/iMnn/4/U7f+Q1u7/jNDq/4PD
	5P9+uOD/gr7j/3634P+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4EzJP+GHwD/b0M+/znU
	5/868P//Oun9/zrr//866///Ouv//zrq//866v//Our//znp//9B7v//F+f7/4TU
	3P/48vH//////7e7v/8ADl7/V32W/5PT7P+IyOr/j9Ps/4fJ5/+Iyej/hMPk/3iu
	2/93rtr/gr/j/3uz3/+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4c/4ExIf+DMCD/gAAA/0ya
	pf8z9f//Pev//zro/f866v//Our//zrq//866v//Oun+/zru//870uP/MXCJ/z5Y
	ef9hY3j/Vlxu/x4kWv9BU3f/isTX/4rM7P+GxeX/j9Xs/4rN6P+GxuX/gsDi/3ar
	2v91qdr/gb3i/3+64f+CMib/gjAh/4IwH/+CMiT/gi8f/4IuHP+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHP+BMiP/hScN/3It
	Hv8+sL7/M/H//zrx//887f//POv//zvr//876v//POv//z32//84x9n/JyZk/0ll
	jP9xqr7/e7nN/4S9zf+Pz+H/kNbw/47R7P+IyOf/icvo/4/U7f+Kzej/iMrm/3iv
	2/92q9n/fLXe/4C64f+DHAD/gioT/4IsGv+BHAD/gSwb/4IyJP+CLhz/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLhz/gjEi/4Un
	D/93HgD/U3Z7/zvAz/8y5Pf/M+v+/zXs//806/7/MuX4/zXj9v8yrLr/TkJK/43C
	0f+f8f//ldv0/47T7/+O1e//iczn/4/U7f+Q1u7/jtLr/47S7P+Hx+b/gLvg/3ux
	2/+CvOH/erDd/3is3P98a4P/gTgt/4Q2H/+GXFv/hDMd/4EfAP+CMyb/gjAg/4Iv
	H/+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4c/4Ex
	Iv+EMSH/hAAA/3cYAP9mUlH/Xmhr/15rb/9eamv/YltZ/2hAO/90MCP/gCUA/4NQ
	Tf+Kq7j/jtbv/4LC5f+DweD/g8Li/4PD4/+O0+v/jtLs/47T7P+Lzen/fbbe/3Wn
	2P+CveL/gLvi/3Sn2f9zq+D/fqHF/36Vuv95q97/g5Ou/4RCOP+BHQD/gisW/4It
	G/+CMSL/gi4d/4IuHP+CLx7/gjAg/4IvHv+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+BLx7/gjUn/4UxIf+GHwD/hhEA/4YPAP+GEAD/hhoA/4YmCf+FLRr/gzEj/4Ic
	AP9+CQD/g3Z5/4vC1/+O2PT/hcTl/3233P+Cv+L/h8jn/4rN6f+Ew+P/hMLi/3es
	2/92rdv/hcPl/3mu3f+TzeP/jsvl/4bA4/92qNn/da/h/4Gs0P+FY2n/gjUm/4I3
	Kv+BJQD/gjIk/4IzJP+CMSL/gSkW/4EuH/+CMSH/gi4c/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi4d/4EvH/+BMiP/gTQm/4E0Jv+BNCb/gTIk/4ExIv+BLx//gi4c/4Iz
	JP+DMiX/gAAA/345Lv+Jnqz/hcTk/3au3/96sNv/iMnn/4bG5v95sdv/frjd/3qw
	3P95rtz/frXf/3Wo2v+W1+r/ldTm/5bV5v+X1OX/k8vg/47N6P+NwNn/ha3M/4Ok
	wP+BTk//gQcA/4EgAP+BHgD/hDch/4QqAP+BJw3/gjIk/4IwIP+CLyD/gi4d/4Iu
	Hf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	HP+CLx7/gjYp/4InEP9/CwD/gGdo/3uhxf9xqeH/dq3b/4C63/98td3/erLc/3ar
	2/9zptn/cKHY/3Gi2P+T2+7/ktjs/5TW6/+V1Oj/l9Tn/5fT5f+W1Ob/kc7j/43O
	6P+Puc3/iF9c/4NVU/+AlKz/gazQ/4aUrf+DPTT/gh8A/4MtF/+CLBb/gjAh/4Iv
	IP+CMCL/gjAg/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4IuHf+CLh3/gi4d/4Iu
	Hf+CLh3/gi0c/4IxIv+CMyb/gRMA/4A0Jv9/iKH/eK3b/3Cm2/9yo9b/d6rb/3Kk
	2P9untf/b53Y/3Ch2f+S3fD/kt7x/5Le8f+S3fD/k9vu/5PY6/+V1er/l9Xn/5fS
	5P+X1un/mdDf/5jM3f+Ozej/f7jh/3u25f9/lbX/fj88/344N/+ANzL/gigK/4Iu
	G/+CKAX/gi0Z/4IyJf+CLx//gi4d/4IuHf+CLh3/gi4d/4IwH/+CMCD/gi8e/4Iu
	Hf+CLh3/gi4d/4IuHf+CLx3/gjQn/4IpEf+CHQD/hWlx/36nzf9wquT/cKHZ/2+f
	1v9vndj/caDZ/26c1/+T3fD/k93v/5Pd7/+S3e//kt3w/5Le8f+R3vH/ktzx/5Pb
	7/+U2Or/l9fp/5fV6f+X0uP/mtHi/5bO4v+OxuH/j7nP/4Ku0v99l7z/gT89/4AZ
	AP9/PkD/gC4f/4MdAP+CLxz/gjAh/4IvH/+CMSP/gjEi/4ItGv+CKhX/gjAg/4Iv
	Hv+CLh3/gi4d/4IuHf+CLh3/gi4c/4IxIv+CMST/gh0A/4I/M/97hKH/bpzV/2ua
	2P9vndf/b53Y/2+d2P+S3O7/ktzv/5Ld8P+T3fD/k93w/5Pd7/+T3e//kt3w/5Ld
	8P+S3vH/k93w/5Lc8P+S2+//lNjs/5XX6f+X1Ob/mNTn/5fQ5P+UzeL/lsLS/5Gz
	yf+Ds9j/gaPD/35gcf99GQD/gSgR/4ItGv+CIwD/gSMA/38wIP+AOSv/gygO/4Ix
	Iv+CMCD/gi0b/4IuHf+CLh3/gi0b/4IxIv+CNCb/gR4A/4BNSP+CgIP/gaG6/3al
	2f9qltX/bZzX/2+d2P+U4PT/kt3x/5Hc7/+R2+3/kdvt/5Hb7v+S3O//ktzv/5Lc
	7/+T3e//kt3v/5Pd8P+S3vD/kt7x/5Pe8P+T3PD/kdrv/5fb7v+OwdL/b5m0/4Cz
	yv+Vy97/ldHn/47L6P+Sqrv/iUk7/4AvJf+ATlL/hVFL/5G6yP+S0eL/hnyA/34A
	AP+DMiT/gjYq/4IzJP+CMyb/gjYq/4MsGv+AAAD/gmhp/5DR5P+U5v3/lt7x/5LT
	6f9/str/bZnV/2uZ1/+O0uX/ltzy/5nj9/+Z5fr/mOX5/5fj+f+V4Pb/k97x/5Lc
	7/+R2+7/kdvu/5Hb7v+S2+7/ktzv/5Pc7/+T3fD/kt3w/5Ld8f+X2Oz/eqzF/zlK
	e/8xQXn/VHCU/26Ys/+Ct8//ja69/5CwwP+Qvtb/ls7g/5Th9P+U5Pn/ktjr/4SE
	if9+AAD/gAAA/4IYAP+BEwD/gAAA/34pD/+Gj5b/ktjr/5Pi9f+S2u3/ktzw/5Pf
	8f+V3vH/isPg/3Ke1P8nMTX/QFRa/1R0ff9mj53/dqi4/4O9z/+NzOD/lNns/5fh
	9P+Z5Pj/mOX5/5bk9/+V4fX/k9/y/5Hd7/+R2+7/kdvu/5HZ7P+U1Oj/mOL1/47N
	4v9ljaz/PVSF/ys2dP8xPnb/U3mb/4W9z/+W2uz/kd7x/5Ha7f+R2ev/kt7x/5Le
	8/+MtcX/g3uB/4BjY/+Aamz/hY6X/4zE1P+T4vX/kt3w/5HZ7P+R2+7/kdvu/5Ha
	7f+R2+7/k97w/47P5v8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8hJCb/OUtQ/09r
	dP9kiJT/dKGw/4G4yf+Myd3/lNbq/5fg8/+Y5Pj/mOb6/5jj9/+Z3PH/luL2/5jo
	/P+c6/3/l+Hz/5LV6v+R1er/lt/x/5fl+P+W5fn/l+P3/5fk9/+X5Pn/mOP6/5jj
	+/+a6v//mur//5jl/f+Z5///muv//5np//+Y4/r/mOP6/5jk/P+Y5Pv/mOT7/5jk
	+/+Y4/v/l+P7/5jm/f8LDxD/DRMV/w8WGP8QGBr/DRUX/wAKDP8AAAD/AAAA/wAA
	AP8AAAD/AAAA/wAAAP8dHh//NkNI/0tjbP9dgo3/bp2q/3yzwv+FuMj/g7vM/4K6
	yv+Busr/g73N/4XA0P+EwND/g73O/4O7y/+Du8v/g7vM/4S8zP+Bucn/ebLC/3ix
	wv94r8D/eLDA/3iywv94scL/eK/A/3iwwP94scL/eLLC/3ixwv94scL/eLHC/3ix
	wv94scL/eLHC/3ixwv8AAAD/AAAA/wAAAP8AAAD/AAAA/wQFBf8JDA3/DRIT/w4V
	F/8PFhj/DBQV/wALDP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAA
	AP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAA
	AP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAA
	AP8AAAD/AAAA/wAAAP8AAAD/AAAA/wEBAf8BAQL/AQIC/wABAf8AAAD/AAAA/wMC
	AP8GBgT/AgID/wUGB/8HCgv/Cw8Q/w4UFv8QFxn/DRYY/wYND/8ABQb/AAgJ/wAI
	Cf8ACAn/AAgJ/wAICf8ACAn/AAgJ/wAHCP8ACAj/AAkK/wAICf8CCwz/CQ8R/wkP
	Ev8JDxH/CQ8R/wkPEv8JDxH/CA4R/xEUFP8WGBb/ExYW/xAUFP8LEBL/CQ8R/wkP
	Ef8HDhH/DhIT/xIVFf8NDAn/Dw4L/wAAAP8CAQH/AgIB/wAAAP8HBwb/CwsJ/wAA
	AP8AAAD/DAwJ/wkIB/8WFA//FhQO/wwKB/8GBAD/FhQO/xIRDv8LDAv/FBMQ/w0O
	DP8JCgr/FxYR/xgXEv8VFRH/FRQQ/xUUEf8SEg//Dg4M/wkKCv8KCwr/AQQE/wAA
	A/8NDQr/FhQP/wsLCf8AAAT/CQkI/wAAAP8AAAD/AAAA/wAAAP8AAAD/BAYG/woK
	Cf8REQ7/AAAA/wAAAP8AAAD/AAAA/xIRDf8AAAD/AAAA/wwLCf8AAAD/AAAA/xUU
	Ef8JCQ3/AAAA/wAAAP8AAAD/AAAA/wAAAP8HBQD/AAAA/wAAAP8AAAD/AAAA/wAA
	AP8HBgD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/AAAA/wAAAP8AAAD/FxUQ/xsZ
	Ev8AAAD/AAAA/wAAAP8YFg7/BwUD/z05LP9WUTv/R0Ez/y8rI/8aGRL/FRMN/wAA
	AP8AAAD/MS8l/z87L/8/Oy7/SkQ1/wAAAP8ODQn/CQkH/wAAAP8YFxT/FRUT/zYy
	Kf95cVv/LCwl/xYWF/9za1H/bmdM/zAtIv8NDxD/dWxS/1FLOv8UFhj/a2JO/zIu
	Jv8AAAz/enBU/3hvUv9eWEL/XFZC/1pUQf9OSDn/hXld/0pEOP9TTD7/AAAA/wAA
	AP9OSDf/npJt/0hCNf8AAAD/AAAA/6OWcP/o1Jf/4c2U/8+9i/9PSDn/AAAA/zEv
	Jv8AAAf/cmlQ/9nHkP+kl3H/s6N5/xwbF/8AAAD/Ih8W/wAAAP+Ad1r/m49s/zIv
	Kf/Ft4X/x7eF/7GieP/m0pf/7tqb/7aoev+Sh2b/5NGW/+HOk/+5q33/lIhs/46C
	Z/+9rX//5tKW/+nVmP/iz5T/4c+U/+HOlP/cyZH/3suS/9C/iv+woXf/dWxQ/21m
	TP+voHT/7tmb/8e2hf+xonn/r6J2/9LBif/m0ZT/5dGU/+XRlP+6rHz/npFr/8u5
	h//Pvor/xLSC/+XRk//jz5T/6dWY/52Ra/8AAAD/AAAA/31zWP/hzZX/28iP/8y7
	h//fzJH/6tWW/+nVl//izpL/4s6S/+bSlv/m0pf/4MyR/+jUlv/hzpT/YVlH/5mM
	aP/w25v/4s6S/+HNkv/jz5P/48+T/+PPk//k0JP/49CT/+fTlf/n0pX/6dWZ/+nV
	mf/l0ZX/4M2R/+jTlf/s15j/69eY/+bRlP/izpL/4c6S/+LOkv/o1Zf/7NiZ/+fT
	lv/m05X/59SW/+HOkv/izpL/5tKV/9nGjv+GfF7/fXVb/9LAi//o1Jb/5NCT/+jT
	lv/jz5P/4M2S/+HNkv/iz5P/4s+T/+LOkv/izpL/4c6S/+PPk//k0ZT/hHpa/7Ok
	dv/q1pj/4c2S/+LPk//izpP/4s6T/+LOk//izpL/4s6S/+HOkv/hzpL/4c6S/+HO
	kv/izpL/4s+T/+HNkv/gzZH/4M2R/+LOkv/iz5P/4s+T/+LPk//hzZL/4M2R/+HO
	kv/hzpL/4c6S/+LPk//izpP/4s6S/+bSlP/hzpX/4s+W/+fTlf/hzZL/4s6T/+HN
	kv/izpP/48+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s6S/+PPk//l0ZX/pphw/8e2
	g//p1Zb/4c6S/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LP
	k//iz5P/4s+T/+LPk//jz5P/48+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LP
	k//iz5P/4s+T/+LPk//iz5P/4s+T/+HOkv/iz5L/4s6S/+HOkv/jz5P/4s+T/+LP
	k//iz5P/4s+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s6S/+PPk//l0ZT/r6F2/869
	iP/o05X/4s6S/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LP
	k//iz5P/4s+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LPk//iz5P/4s+T/+LP
	k//iz5P/4s+T/+LPk/8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	AAAAAAAAAAAAAAAAAAA=
`