"use strict";
(function() {

Error.stackTraceLimit = -1;

var go$reservedKeywords = ["abstract", "arguments", "boolean", "break", "byte", "case", "catch", "char", "class", "const", "continue", "debugger", "default", "delete", "do", "double", "else", "enum", "eval", "export", "extends", "false", "final", "finally", "float", "for", "function", "goto", "if", "implements", "import", "in", "instanceof", "int", "interface", "let", "long", "native", "new", "package", "private", "protected", "public", "return", "short", "static", "super", "switch", "synchronized", "this", "throw", "throws", "transient", "true", "try", "typeof", "var", "void", "volatile", "while", "with", "yield"];

var go$global;
if (typeof window !== "undefined") {
	go$global = window;
} else if (typeof GLOBAL !== "undefined") {
	go$global = GLOBAL;
	go$global.require = require;
}

var go$idCounter = 0;
var go$keys = function(m) { return m ? Object.keys(m) : []; };
var go$min = Math.min;
var go$parseInt = parseInt;
var go$parseFloat = function(f) {
	if (f.constructor === Number) {
		return f;
	}
	return parseFloat(f);
};
var go$mod = function(x, y) { return x % y; };
var go$toString = String;
var go$reflect, go$newStringPtr;
var Go$Array = Array;

var go$floatKey = function(f) {
	if (f !== f) {
		go$idCounter++;
		return "NaN$" + go$idCounter;
	}
	return String(f);
};

var go$mapArray = function(array, f) {
	var newArray = new array.constructor(array.length), i;
	for (i = 0; i < array.length; i++) {
		newArray[i] = f(array[i]);
	}
	return newArray;
};

var go$newType = function(size, kind, string, name, pkgPath, constructor) {
	var typ;
	switch(kind) {
	case "Bool":
	case "Int":
	case "Int8":
	case "Int16":
	case "Int32":
	case "Uint":
	case "Uint8" :
	case "Uint16":
	case "Uint32":
	case "Uintptr":
	case "String":
	case "UnsafePointer":
		typ = function(v) { this.go$val = v; };
		typ.prototype.go$key = function() { return string + "$" + this.go$val; };
		break;

	case "Float32":
	case "Float64":
		typ = function(v) { this.go$val = v; };
		typ.prototype.go$key = function() { return string + "$" + go$floatKey(this.go$val); };
		break;

	case "Int64":
		typ = function(high, low) {
			this.high = (high + Math.floor(Math.ceil(low) / 4294967296)) >> 0;
			this.low = low >>> 0;
			this.go$val = this;
		};
		typ.prototype.go$key = function() { return string + "$" + this.high + "$" + this.low; };
		break;

	case "Uint64":
		typ = function(high, low) {
			this.high = (high + Math.floor(Math.ceil(low) / 4294967296)) >>> 0;
			this.low = low >>> 0;
			this.go$val = this;
		};
		typ.prototype.go$key = function() { return string + "$" + this.high + "$" + this.low; };
		break;

	case "Complex64":
	case "Complex128":
		typ = function(real, imag) {
			this.real = real;
			this.imag = imag;
			this.go$val = this;
		};
		typ.prototype.go$key = function() { return string + "$" + this.real + "$" + this.imag; };
		break;

	case "Array":
		typ = function(v) { this.go$val = v; };
		typ.Ptr = go$newType(4, "Ptr", "*" + string, "", "", function(array) {
			this.go$get = function() { return array; };
			this.go$val = array;
		});
		typ.init = function(elem, len) {
			typ.elem = elem;
			typ.len = len;
			typ.prototype.go$key = function() {
				return string + "$" + go$mapArray(this.go$val, function(e) {
					var key = e.go$key ? e.go$key() : String(e);
					return key.replace(/\\/g, "\\\\").replace(/\$/g, "\\$");
				}).join("$");
			};
			typ.extendReflectType = function(rt) {
				rt.arrayType = new go$reflect.arrayType(rt, elem.reflectType(), undefined, len);
			};
			typ.Ptr.init(typ);
		};
		break;

	case "Chan":
		typ = function() { this.go$val = this; };
		typ.prototype.go$key = function() {
			if (this.go$id === undefined) {
				go$idCounter++;
				this.go$id = go$idCounter;
			}
			return String(this.go$id);
		};
		typ.init = function(elem, sendOnly, recvOnly) {
			typ.nil = new typ();
			typ.extendReflectType = function(rt) {
				rt.chanType = new go$reflect.chanType(rt, elem.reflectType(), sendOnly ? go$reflect.SendDir : (recvOnly ? go$reflect.RecvDir : go$reflect.BothDir));
			};
		};
		break;

	case "Func":
		typ = function(v) { this.go$val = v; };
		typ.init = function(params, results, variadic) {
			typ.params = params;
			typ.results = results;
			typ.variadic = variadic;
			typ.extendReflectType = function(rt) {
				var typeSlice = (go$sliceType(go$ptrType(go$reflect.rtype)));
				rt.funcType = new go$reflect.funcType(rt, variadic, new typeSlice(go$mapArray(params, function(p) { return p.reflectType(); })), new typeSlice(go$mapArray(results, function(p) { return p.reflectType(); })));
			};
		};
		break;

	case "Interface":
		typ = { implementedBy: [] };
		typ.init = function(methods) {
			typ.extendReflectType = function(rt) {
				var imethods = go$mapArray(methods, function(m) {
					return new go$reflect.imethod(go$newStringPtr(m[0]), go$newStringPtr(m[1]), m[2].reflectType());
				});
				var methodSlice = (go$sliceType(go$ptrType(go$reflect.imethod)));
				rt.interfaceType = new go$reflect.interfaceType(rt, new methodSlice(imethods));
			};
		};
		break;

	case "Map":
		typ = function(v) { this.go$val = v; };
		typ.init = function(key, elem) {
			typ.key = key;
			typ.elem = elem;
			typ.extendReflectType = function(rt) {
				rt.mapType = new go$reflect.mapType(rt, key.reflectType(), elem.reflectType(), undefined, undefined);
			};
		};
		break;

	case "Ptr":
		typ = constructor || function(getter, setter) {
			this.go$get = getter;
			this.go$set = setter;
			this.go$val = this;
		};
		typ.prototype.go$key = function() {
			if (this.go$id === undefined) {
				go$idCounter++;
				this.go$id = go$idCounter;
			}
			return String(this.go$id);
		};
		typ.init = function(elem) {
			typ.nil = new typ(go$throwNilPointerError, go$throwNilPointerError);
			typ.extendReflectType = function(rt) {
				rt.ptrType = new go$reflect.ptrType(rt, elem.reflectType());
			};
		};
		break;

	case "Slice":
		var nativeArray;
		typ = function(array) {
			if (array.constructor !== nativeArray) {
				array = new nativeArray(array);
			}
			this.array = array;
			this.offset = 0;
			this.length = array.length;
			this.capacity = array.length;
			this.go$val = this;
		};
		typ.make = function(length, capacity, zero) {
			capacity = capacity || length;
			var array = new nativeArray(capacity), i;
			for (i = 0; i < capacity; i++) {
				array[i] = zero();
			}
			var slice = new typ(array);
			slice.length = length;
			return slice;
		};
		typ.init = function(elem) {
			typ.elem = elem;
			nativeArray = go$nativeArray(elem.kind);
			typ.nil = new typ([]);
			typ.extendReflectType = function(rt) {
				rt.sliceType = new go$reflect.sliceType(rt, elem.reflectType());
			};
		};
		break;

	case "Struct":
		typ = function(v) { this.go$val = v; };
		typ.Ptr = go$newType(4, "Ptr", "*" + string, "", "", constructor);
		typ.Ptr.Struct = typ;
		typ.init = function(fields) {
			var i;
			typ.fields = fields;
			typ.Ptr.init(typ);
			// nil value
			typ.Ptr.nil = new constructor();
			for (i = 0; i < fields.length; i++) {
				var field = fields[i];
				Object.defineProperty(typ.Ptr.nil, field[1], { get: go$throwNilPointerError, set: go$throwNilPointerError });
			}
			// methods for embedded fields
			for (i = 0; i < typ.methods.length; i++) {
				var method = typ.methods[i];
				if (method[5] != -1) {
					(function(field, methodName) {
						typ.prototype[methodName] = function() {
							var v = this.go$val[field[0]];
							return v[methodName].apply(v, arguments);
						};
					})(fields[method[5]], method[0]);
				}
			}
			for (i = 0; i < typ.Ptr.methods.length; i++) {
				var method = typ.Ptr.methods[i];
				if (method[5] != -1) {
					(function(field, methodName) {
						typ.Ptr.prototype[methodName] = function() {
							var v = this[field[0]];
							if (v.go$val === undefined) {
								v = new field[3](v);
							}
							return v[methodName].apply(v, arguments);
						};
					})(fields[method[5]], method[0]);
				}
			}
			// map key
			typ.prototype.go$key = function() {
				var keys = new Array(fields.length);
				for (i = 0; i < fields.length; i++) {
					var v = this.go$val[fields[i][0]];
					var key = v.go$key ? v.go$key() : String(v);
					keys[i] = key.replace(/\\/g, "\\\\").replace(/\$/g, "\\$");
				}
				return string + "$" + keys.join("$");
			};
			// reflect type
			typ.extendReflectType = function(rt) {
				var reflectFields = new Array(fields.length), i;
				for (i = 0; i < fields.length; i++) {
					var field = fields[i];
					reflectFields[i] = new go$reflect.structField(go$newStringPtr(field[1]), go$newStringPtr(field[2]), field[3].reflectType(), go$newStringPtr(field[4]), i);
				}
				rt.structType = new go$reflect.structType(rt, new (go$sliceType(go$reflect.structField))(reflectFields));
			};
		};
		break;

	default:
		throw go$panic(new Go$String("invalid kind: " + kind));
	}

	typ.kind = kind;
	typ.string = string;
	typ.typeName = name;
	typ.pkgPath = pkgPath;
	typ.methods = [];
	var rt = null;
	typ.reflectType = function() {
		if (rt === null) {
			rt = new go$reflect.rtype(size, 0, 0, 0, 0, go$reflect.kinds[kind], undefined, undefined, go$newStringPtr(string), undefined, undefined);
			rt.jsType = typ;

			var methods = [];
			if (typ.methods !== undefined) {
				var i;
				for (i = 0; i < typ.methods.length; i++) {
					var m = typ.methods[i];
					methods.push(new go$reflect.method(go$newStringPtr(m[0]), go$newStringPtr(m[1]), go$funcType(m[2], m[3], m[4]).reflectType(), go$funcType([typ].concat(m[2]), m[3], m[4]).reflectType(), undefined, undefined));
				}
			}
			if (name !== "" || methods.length !== 0) {
				var methodSlice = (go$sliceType(go$ptrType(go$reflect.method)));
				rt.uncommonType = new go$reflect.uncommonType(go$newStringPtr(name), go$newStringPtr(pkgPath), new methodSlice(methods));
			}

			if (typ.extendReflectType !== undefined) {
				typ.extendReflectType(rt);
			}
		}
		return rt;
	};
	return typ;
};

var Go$Bool          = go$newType( 1, "Bool",          "bool",           "bool",       "", null);
var Go$Int           = go$newType( 4, "Int",           "int",            "int",        "", null);
var Go$Int8          = go$newType( 1, "Int8",          "int8",           "int8",       "", null);
var Go$Int16         = go$newType( 2, "Int16",         "int16",          "int16",      "", null);
var Go$Int32         = go$newType( 4, "Int32",         "int32",          "int32",      "", null);
var Go$Int64         = go$newType( 8, "Int64",         "int64",          "int64",      "", null);
var Go$Uint          = go$newType( 4, "Uint",          "uint",           "uint",       "", null);
var Go$Uint8         = go$newType( 1, "Uint8",         "uint8",          "uint8",      "", null);
var Go$Uint16        = go$newType( 2, "Uint16",        "uint16",         "uint16",     "", null);
var Go$Uint32        = go$newType( 4, "Uint32",        "uint32",         "uint32",     "", null);
var Go$Uint64        = go$newType( 8, "Uint64",        "uint64",         "uint64",     "", null);
var Go$Uintptr       = go$newType( 4, "Uintptr",       "uintptr",        "uintptr",    "", null);
var Go$Float32       = go$newType( 4, "Float32",       "float32",        "float32",    "", null);
var Go$Float64       = go$newType( 8, "Float64",       "float64",        "float64",    "", null);
var Go$Complex64     = go$newType( 8, "Complex64",     "complex64",      "complex64",  "", null);
var Go$Complex128    = go$newType(16, "Complex128",    "complex128",     "complex128", "", null);
var Go$String        = go$newType( 8, "String",        "string",         "string",     "", null);
var Go$UnsafePointer = go$newType( 4, "UnsafePointer", "unsafe.Pointer", "Pointer",    "", null);

var go$nativeArray = function(elemKind) {
	return ({ Int: Int32Array, Int8: Int8Array, Int16: Int16Array, Int32: Int32Array, Uint: Uint32Array, Uint8: Uint8Array, Uint16: Uint16Array, Uint32: Uint32Array, Uintptr: Uint32Array, Float32: Float32Array, Float64: Float64Array })[elemKind] || Array;
};
var go$toNativeArray = function(elemKind, array) {
	var nativeArray = go$nativeArray(elemKind);
	if (nativeArray === Array) {
		return array;
	}
	return new nativeArray(array);
};
var go$makeNativeArray = function(elemKind, length, zero) {
	var array = new (go$nativeArray(elemKind))(length), i;
	for (i = 0; i < length; i++) {
		array[i] = zero();
	}
	return array;
};
var go$arrayTypes = {};
var go$arrayType = function(elem, len) {
	var string = "[" + len + "]" + elem.string;
	var typ = go$arrayTypes[string];
	if (typ === undefined) {
		typ = go$newType(12, "Array", string, "", "", null);
		typ.init(elem, len);
		go$arrayTypes[string] = typ;
	}
	return typ;
};

var go$chanType = function(elem, sendOnly, recvOnly) {
	var string = (recvOnly ? "<-" : "") + "chan" + (sendOnly ? "<- " : " ") + elem.string;
	var field = sendOnly ? "SendChan" : (recvOnly ? "RecvChan" : "Chan");
	var typ = elem[field];
	if (typ === undefined) {
		typ = go$newType(4, "Chan", string, "", "", null);
		typ.init(elem, sendOnly, recvOnly);
		elem[field] = typ;
	}
	return typ;
};

var go$funcTypes = {};
var go$funcType = function(params, results, variadic) {
	var paramTypes = go$mapArray(params, function(p) { return p.string; });
	if (variadic) {
		paramTypes[paramTypes.length - 1] = "..." + paramTypes[paramTypes.length - 1].substr(2);
	}
	var string = "func(" + paramTypes.join(", ") + ")";
	if (results.length === 1) {
		string += " " + results[0].string;
	} else if (results.length > 1) {
		string += " (" + go$mapArray(results, function(r) { return r.string; }).join(", ") + ")";
	}
	var typ = go$funcTypes[string];
	if (typ === undefined) {
		typ = go$newType(4, "Func", string, "", "", null);
		typ.init(params, results, variadic);
		go$funcTypes[string] = typ;
	}
	return typ;
};

var go$interfaceTypes = {};
var go$interfaceType = function(methods) {
	var string = "interface {}";
	if (methods.length !== 0) {
		string = "interface { " + go$mapArray(methods, function(m) {
			return (m[1] !== "" ? m[1] + "." : "") + m[0] + m[2].string.substr(4);
		}).join("; ") + " }";
	}
	var typ = go$interfaceTypes[string];
	if (typ === undefined) {
		typ = go$newType(8, "Interface", string, "", "", null);
		typ.init(methods);
		go$interfaceTypes[string] = typ;
	}
	return typ;
};
var go$emptyInterface = go$interfaceType([]);
var go$interfaceNil = { go$key: function() { return "nil"; } };
var go$error = go$newType(8, "Interface", "error", "error", "", null);
go$error.init([["Error", "", go$funcType([], [Go$String], false)]]);

var Go$Map = function() {};
(function() {
	var names = Object.getOwnPropertyNames(Object.prototype), i;
	for (i = 0; i < names.length; i++) {
		Go$Map.prototype[names[i]] = undefined;
	}
})();
var go$mapTypes = {};
var go$mapType = function(key, elem) {
	var string = "map[" + key.string + "]" + elem.string;
	var typ = go$mapTypes[string];
	if (typ === undefined) {
		typ = go$newType(4, "Map", string, "", "", null);
		typ.init(key, elem);
		go$mapTypes[string] = typ;
	}
	return typ;
};

var go$throwNilPointerError = function() { go$throwRuntimeError("invalid memory address or nil pointer dereference"); };
var go$ptrType = function(elem) {
	var typ = elem.Ptr;
	if (typ === undefined) {
		typ = go$newType(4, "Ptr", "*" + elem.string, "", "", null);
		typ.init(elem);
		elem.Ptr = typ;
	}
	return typ;
};

var go$sliceType = function(elem) {
	var typ = elem.Slice;
	if (typ === undefined) {
		typ = go$newType(12, "Slice", "[]" + elem.string, "", "", null);
		typ.init(elem);
		elem.Slice = typ;
	}
	return typ;
};

var go$structTypes = {};
var go$structType = function(fields) {
	var string = "struct { " + go$mapArray(fields, function(f) {
		return f[1] + " " + f[3].string + (f[4] !== "" ? (' "' + f[4].replace(/\\/g, "\\\\").replace(/"/g, '\\"') + '"') : "");
	}).join("; ") + " }";
	var typ = go$structTypes[string];
	if (typ === undefined) {
		typ = go$newType(0, "Struct", string, "", "", function() {
			this.go$val = this;
			var i;
			for (i = 0; i < fields.length; i++) {
				this[fields[i][0]] = arguments[i];
			}
		});
		typ.init(fields);
		go$structTypes[string] = typ;
	}
	return typ;
};

var go$stringPtrMap = new Go$Map();
go$newStringPtr = function(str) {
	if (str === undefined || str === "") {
		return go$ptrType(Go$String).nil;
	}
	var ptr = go$stringPtrMap[str];
	if (ptr === undefined) {
		ptr = new (go$ptrType(Go$String))(function() { return str; }, function(v) { str = v; });
		go$stringPtrMap[str] = ptr;
	}
	return ptr;
};
var go$newDataPointer = function(data, constructor) {
	return new constructor(function() { return data; }, function(v) { data = v; });
};

var go$coerceFloat32 = function(f) {
	var math = go$packages["math"];
	if (math === undefined) {
		return f;
	}
	return math.Float32frombits(math.Float32bits(f));
};
var go$flatten64 = function(x) {
	return x.high * 4294967296 + x.low;
};
var go$shiftLeft64 = function(x, y) {
	if (y === 0) {
		return x;
	}
	if (y < 32) {
		return new x.constructor(x.high << y | x.low >>> (32 - y), (x.low << y) >>> 0);
	}
	if (y < 64) {
		return new x.constructor(x.low << (y - 32), 0);
	}
	return new x.constructor(0, 0);
};
var go$shiftRightInt64 = function(x, y) {
	if (y === 0) {
		return x;
	}
	if (y < 32) {
		return new x.constructor(x.high >> y, (x.low >>> y | x.high << (32 - y)) >>> 0);
	}
	if (y < 64) {
		return new x.constructor(x.high >> 31, (x.high >> (y - 32)) >>> 0);
	}
	if (x.high < 0) {
		return new x.constructor(-1, 4294967295);
	}
	return new x.constructor(0, 0);
};
var go$shiftRightUint64 = function(x, y) {
	if (y === 0) {
		return x;
	}
	if (y < 32) {
		return new x.constructor(x.high >>> y, (x.low >>> y | x.high << (32 - y)) >>> 0);
	}
	if (y < 64) {
		return new x.constructor(0, x.high >>> (y - 32));
	}
	return new x.constructor(0, 0);
};
var go$mul64 = function(x, y) {
	var high = 0, low = 0, i;
	if ((y.low & 1) !== 0) {
		high = x.high;
		low = x.low;
	}
	for (i = 1; i < 32; i++) {
		if ((y.low & 1<<i) !== 0) {
			high += x.high << i | x.low >>> (32 - i);
			low += (x.low << i) >>> 0;
		}
	}
	for (i = 0; i < 32; i++) {
		if ((y.high & 1<<i) !== 0) {
			high += x.low << i;
		}
	}
	return new x.constructor(high, low);
};
var go$div64 = function(x, y, returnRemainder) {
	if (y.high === 0 && y.low === 0) {
		go$throwRuntimeError("integer divide by zero");
	}

	var s = 1;
	var rs = 1;

	var xHigh = x.high;
	var xLow = x.low;
	if (xHigh < 0) {
		s = -1;
		rs = -1;
		xHigh = -xHigh;
		if (xLow !== 0) {
			xHigh--;
			xLow = 4294967296 - xLow;
		}
	}

	var yHigh = y.high;
	var yLow = y.low;
	if (y.high < 0) {
		s *= -1;
		yHigh = -yHigh;
		if (yLow !== 0) {
			yHigh--;
			yLow = 4294967296 - yLow;
		}
	}

	var high = 0, low = 0, n = 0, i;
	while (yHigh < 2147483648 && ((xHigh > yHigh) || (xHigh === yHigh && xLow > yLow))) {
		yHigh = (yHigh << 1 | yLow >>> 31) >>> 0;
		yLow = (yLow << 1) >>> 0;
		n++;
	}
	for (i = 0; i <= n; i++) {
		high = high << 1 | low >>> 31;
		low = (low << 1) >>> 0;
		if ((xHigh > yHigh) || (xHigh === yHigh && xLow >= yLow)) {
			xHigh = xHigh - yHigh;
			xLow = xLow - yLow;
			if (xLow < 0) {
				xHigh--;
				xLow += 4294967296;
			}
			low++;
			if (low === 4294967296) {
				high++;
				low = 0;
			}
		}
		yLow = (yLow >>> 1 | yHigh << (32 - 1)) >>> 0;
		yHigh = yHigh >>> 1;
	}

	if (returnRemainder) {
		return new x.constructor(xHigh * rs, xLow * rs);
	}
	return new x.constructor(high * s, low * s);
};

var go$divComplex = function(n, d) {
	var ninf = n.real === 1/0 || n.real === -1/0 || n.imag === 1/0 || n.imag === -1/0;
	var dinf = d.real === 1/0 || d.real === -1/0 || d.imag === 1/0 || d.imag === -1/0;
	var nnan = !ninf && (n.real !== n.real || n.imag !== n.imag);
	var dnan = !dinf && (d.real !== d.real || d.imag !== d.imag);
	if(nnan || dnan) {
		return new n.constructor(0/0, 0/0);
	}
	if (ninf && !dinf) {
		return new n.constructor(1/0, 1/0);
	}
	if (!ninf && dinf) {
		return new n.constructor(0, 0);
	}
	if (d.real === 0 && d.imag === 0) {
		if (n.real === 0 && n.imag === 0) {
			return new n.constructor(0/0, 0/0);
		}
		return new n.constructor(1/0, 1/0);
	}
	var a = Math.abs(d.real);
	var b = Math.abs(d.imag);
	if (a <= b) {
		var ratio = d.real / d.imag;
		var denom = d.real * ratio + d.imag;
		return new n.constructor((n.real * ratio + n.imag) / denom, (n.imag * ratio - n.real) / denom);
	}
	var ratio = d.imag / d.real;
	var denom = d.imag * ratio + d.real;
	return new n.constructor((n.imag * ratio + n.real) / denom, (n.imag - n.real * ratio) / denom);
};

var go$subslice = function(slice, low, high, max) {
	if (low < 0 || high < low || max < high || high > slice.capacity || max > slice.capacity) {
		go$throwRuntimeError("slice bounds out of range");
	}
	var s = new slice.constructor(slice.array);
	s.offset = slice.offset + low;
	s.length = slice.length - low;
	s.capacity = slice.capacity - low;
	if (high !== undefined) {
		s.length = high - low;
	}
	if (max !== undefined) {
		s.capacity = max - low;
	}
	return s;
};

var go$sliceToArray = function(slice) {
	if (slice.length === 0) {
		return [];
	}
	if (slice.array.constructor !== Array) {
		return slice.array.subarray(slice.offset, slice.offset + slice.length);
	}
	return slice.array.slice(slice.offset, slice.offset + slice.length);
};

var go$decodeRune = function(str, pos) {
	var c0 = str.charCodeAt(pos);

	if (c0 < 0x80) {
		return [c0, 1];
	}

	if (c0 !== c0 || c0 < 0xC0) {
		return [0xFFFD, 1];
	}

	var c1 = str.charCodeAt(pos + 1);
	if (c1 !== c1 || c1 < 0x80 || 0xC0 <= c1) {
		return [0xFFFD, 1];
	}

	if (c0 < 0xE0) {
		var r = (c0 & 0x1F) << 6 | (c1 & 0x3F);
		if (r <= 0x7F) {
			return [0xFFFD, 1];
		}
		return [r, 2];
	}

	var c2 = str.charCodeAt(pos + 2);
	if (c2 !== c2 || c2 < 0x80 || 0xC0 <= c2) {
		return [0xFFFD, 1];
	}

	if (c0 < 0xF0) {
		var r = (c0 & 0x0F) << 12 | (c1 & 0x3F) << 6 | (c2 & 0x3F);
		if (r <= 0x7FF) {
			return [0xFFFD, 1];
		}
		if (0xD800 <= r && r <= 0xDFFF) {
			return [0xFFFD, 1];
		}
		return [r, 3];
	}

	var c3 = str.charCodeAt(pos + 3);
	if (c3 !== c3 || c3 < 0x80 || 0xC0 <= c3) {
		return [0xFFFD, 1];
	}

	if (c0 < 0xF8) {
		var r = (c0 & 0x07) << 18 | (c1 & 0x3F) << 12 | (c2 & 0x3F) << 6 | (c3 & 0x3F);
		if (r <= 0xFFFF || 0x10FFFF < r) {
			return [0xFFFD, 1];
		}
		return [r, 4];
	}

	return [0xFFFD, 1];
};

var go$encodeRune = function(r) {
	if (r < 0 || r > 0x10FFFF || (0xD800 <= r && r <= 0xDFFF)) {
		r = 0xFFFD;
	}
	if (r <= 0x7F) {
		return String.fromCharCode(r);
	}
	if (r <= 0x7FF) {
		return String.fromCharCode(0xC0 | r >> 6, 0x80 | (r & 0x3F));
	}
	if (r <= 0xFFFF) {
		return String.fromCharCode(0xE0 | r >> 12, 0x80 | (r >> 6 & 0x3F), 0x80 | (r & 0x3F));
	}
	return String.fromCharCode(0xF0 | r >> 18, 0x80 | (r >> 12 & 0x3F), 0x80 | (r >> 6 & 0x3F), 0x80 | (r & 0x3F));
};

var go$stringToBytes = function(str, terminateWithNull) {
	var array = new Uint8Array(terminateWithNull ? str.length + 1 : str.length), i;
	for (i = 0; i < str.length; i++) {
		array[i] = str.charCodeAt(i);
	}
	if (terminateWithNull) {
		array[str.length] = 0;
	}
	return array;
};

var go$bytesToString = function(slice) {
	if (slice.length === 0) {
		return "";
	}
	var str = "", i;
	for (i = 0; i < slice.length; i += 10000) {
		str += String.fromCharCode.apply(null, slice.array.subarray(slice.offset + i, slice.offset + Math.min(slice.length, i + 10000)));
	}
	return str;
};

var go$stringToRunes = function(str) {
	var array = new Int32Array(str.length);
	var rune, i, j = 0;
	for (i = 0; i < str.length; i += rune[1], j++) {
		rune = go$decodeRune(str, i);
		array[j] = rune[0];
	}
	return array.subarray(0, j);
};

var go$runesToString = function(slice) {
	if (slice.length === 0) {
		return "";
	}
	var str = "", i;
	for (i = 0; i < slice.length; i++) {
		str += go$encodeRune(slice.array[slice.offset + i]);
	}
	return str;
};

var go$needsExternalization = function(t) {
	switch (t.kind) {
		case "Int64":
		case "Uint64":
		case "Array":
		case "Func":
		case "Interface":
		case "Map":
		case "Slice":
		case "String":
			return true;
		default:
			return false;
	}
};

var go$externalize = function(v, t) {
	switch (t.kind) {
	case "Int64":
	case "Uint64":
		return go$flatten64(v);
	case "Array":
		if (go$needsExternalization(t.elem)) {
			return go$mapArray(v, function(e) { return go$externalize(e, t.elem); });
		}
		return v;
	case "Func":
		if (v === go$throwNilPointerError) {
			return null;
		}
		var convert = false;
		var i;
		for (i = 0; i < t.params.length; i++) {
			convert = convert || (t.params[i] !== go$packages["github.com/gopherjs/gopherjs/js"].Object);
		}
		for (i = 0; i < t.results.length; i++) {
			convert = convert || go$needsExternalization(t.results[i]);
		}
		if (!convert) {
			return v;
		}
		return function() {
			var args = [], i;
			for (i = 0; i < t.params.length; i++) {
				if (t.variadic && i === t.params.length - 1) {
					var vt = t.params[i].elem, varargs = [], j;
					for (j = i; j < arguments.length; j++) {
						varargs.push(go$internalize(arguments[j], vt));
					}
					args.push(new (t.params[i])(varargs));
					break;
				}
				args.push(go$internalize(arguments[i], t.params[i]));
			}
			var result = v.apply(undefined, args);
			switch (t.results.length) {
			case 0:
				return;
			case 1:
				return go$externalize(result, t.results[0]);
			default:
				for (i = 0; i < t.results.length; i++) {
					result[i] = go$externalize(result[i], t.results[i]);
				}
				return result;
			}
		};
	case "Interface":
		if (v === null) {
			return null;
		}
		if (v.constructor.kind === undefined) {
			return v; // js.Object
		}
		return go$externalize(v.go$val, v.constructor);
	case "Map":
		var m = {};
		var keys = go$keys(v), i;
		for (i = 0; i < keys.length; i++) {
			var entry = v[keys[i]];
			m[go$externalize(entry.k, t.key)] = go$externalize(entry.v, t.elem);
		}
		return m;
	case "Slice":
		if (go$needsExternalization(t.elem)) {
			return go$mapArray(go$sliceToArray(v), function(e) { return go$externalize(e, t.elem); });
		}
		return go$sliceToArray(v);
	case "String":
		var s = "", r, i, j = 0;
		for (i = 0; i < v.length; i += r[1], j++) {
			r = go$decodeRune(v, i);
			s += String.fromCharCode(r[0]);
		}
		return s;
	case "Struct":
		var timePkg = go$packages["time"];
		if (timePkg && v.constructor === timePkg.Time.Ptr) {
			var milli = go$div64(v.UnixNano(), new Go$Int64(0, 1000000));
			return new Date(go$flatten64(milli));
		}
		return v;
	default:
		return v;
	}
};

var go$internalize = function(v, t, recv) {
	switch (t.kind) {
	case "Bool":
		return !!v;
	case "Int":
		return parseInt(v);
	case "Int8":
		return parseInt(v) << 24 >> 24;
	case "Int16":
		return parseInt(v) << 16 >> 16;
	case "Int32":
		return parseInt(v) >> 0;
	case "Uint":
		return parseInt(v);
	case "Uint8" :
		return parseInt(v) << 24 >>> 24;
	case "Uint16":
		return parseInt(v) << 16 >>> 16;
	case "Uint32":
	case "Uintptr":
		return parseInt(v) >>> 0;
	case "Int64":
	case "Uint64":
		return new t(0, v);
	case "Float32":
	case "Float64":
		return parseFloat(v);
	case "Array":
		if (v.length !== t.len) {
			go$throwRuntimeError("got array with wrong size from JavaScript native");
		}
		return go$mapArray(v, function(e) { return go$internalize(e, t.elem); });
	case "Func":
		return function() {
			var args = [], i;
			for (i = 0; i < t.params.length; i++) {
				if (t.variadic && i === t.params.length - 1) {
					var vt = t.params[i].elem, varargs = arguments[i], j;
					for (j = 0; j < varargs.length; j++) {
						args.push(go$externalize(varargs.array[varargs.offset + j], vt));
					}
					break;
				}
				args.push(go$externalize(arguments[i], t.params[i]));
			}
			var result = v.apply(recv, args);
			switch (t.results.length) {
			case 0:
				return;
			case 1:
				return go$internalize(result, t.results[0]);
			default:
				for (i = 0; i < t.results.length; i++) {
					result[i] = go$internalize(result[i], t.results[i]);
				}
				return result;
			}
		};
	case "Interface":
		if (t === go$packages["github.com/gopherjs/gopherjs/js"].Object) {
			return v;
		}
		switch (v.constructor) {
		case Int8Array:
			return new (go$sliceType(Go$Int8))(v);
		case Int16Array:
			return new (go$sliceType(Go$Int16))(v);
		case Int32Array:
			return new (go$sliceType(Go$Int))(v);
		case Uint8Array:
			return new (go$sliceType(Go$Uint8))(v);
		case Uint16Array:
			return new (go$sliceType(Go$Uint16))(v);
		case Uint32Array:
			return new (go$sliceType(Go$Uint))(v);
		case Float32Array:
			return new (go$sliceType(Go$Float32))(v);
		case Float64Array:
			return new (go$sliceType(Go$Float64))(v);
		case Array:
			return go$internalize(v, go$sliceType(go$emptyInterface));
		case Boolean:
			return new Go$Bool(!!v);
		case Date:
			var timePkg = go$packages["time"];
			if (timePkg) {
				return new timePkg.Time(timePkg.Unix(new Go$Int64(0, 0), new Go$Int64(0, v.getTime() * 1000000)));
			}
		case Function:
			var funcType = go$funcType([go$sliceType(go$emptyInterface)], [go$packages["github.com/gopherjs/gopherjs/js"].Object], true);
			return new funcType(go$internalize(v, funcType));
		case Number:
			return new Go$Float64(parseFloat(v));
		case Object:
			var mapType = go$mapType(Go$String, go$emptyInterface);
			return new mapType(go$internalize(v, mapType));
		case String:
			return new Go$String(go$internalize(v, Go$String));
		}
		return v;
	case "Map":
		var m = new Go$Map();
		var keys = go$keys(v), i;
		for (i = 0; i < keys.length; i++) {
			var key = go$internalize(keys[i], t.key);
			m[key.go$key ? key.go$key() : key] = { k: key, v: go$internalize(v[keys[i]], t.elem) };
		}
		return m;
	case "Slice":
		return new t(go$mapArray(v, function(e) { return go$internalize(e, t.elem); }));
	case "String":
		v = String(v);
		var s = "", i;
		for (i = 0; i < v.length; i++) {
			s += go$encodeRune(v.charCodeAt(i));
		}
		return s;
	default:
		return v;
	}
};

var go$copySlice = function(dst, src) {
	var n = Math.min(src.length, dst.length), i;
	if (dst.array.constructor !== Array && n !== 0) {
		dst.array.set(src.array.subarray(src.offset, src.offset + n), dst.offset);
		return n;
	}
	for (i = 0; i < n; i++) {
		dst.array[dst.offset + i] = src.array[src.offset + i];
	}
	return n;
};

var go$copyString = function(dst, src) {
	var n = Math.min(src.length, dst.length), i;
	for (i = 0; i < n; i++) {
		dst.array[dst.offset + i] = src.charCodeAt(i);
	}
	return n;
};

var go$copyArray = function(dst, src) {
	var i;
	for (i = 0; i < src.length; i++) {
		dst[i] = src[i];
	}
};

var go$growSlice = function(slice, length) {
	var newCapacity = Math.max(length, slice.capacity < 1024 ? slice.capacity * 2 : Math.floor(slice.capacity * 5 / 4));

	var newArray;
	if (slice.array.constructor === Array) {
		newArray = slice.array;
		if (slice.offset !== 0 || newArray.length !== slice.offset + slice.capacity) {
			newArray = newArray.slice(slice.offset);
		}
		newArray.length = newCapacity;
	} else {
		newArray = new slice.array.constructor(newCapacity);
		newArray.set(slice.array.subarray(slice.offset));
	}

	var newSlice = new slice.constructor(newArray);
	newSlice.length = slice.length;
	newSlice.capacity = newCapacity;
	return newSlice;
};

var go$append = function(slice) {
	if (arguments.length === 1) {
		return slice;
	}

	var newLength = slice.length + arguments.length - 1;
	if (newLength > slice.capacity) {
		slice = go$growSlice(slice, newLength);
	}

	var array = slice.array;
	var leftOffset = slice.offset + slice.length - 1, i;
	for (i = 1; i < arguments.length; i++) {
		array[leftOffset + i] = arguments[i];
	}

	var newSlice = new slice.constructor(array);
	newSlice.offset = slice.offset;
	newSlice.length = newLength;
	newSlice.capacity = slice.capacity;
	return newSlice;
};

var go$appendSlice = function(slice, toAppend) {
	if (toAppend.length === 0) {
		return slice;
	}

	var newLength = slice.length + toAppend.length;
	if (newLength > slice.capacity) {
		slice = go$growSlice(slice, newLength);
	}

	var array = slice.array;
	var leftOffset = slice.offset + slice.length, rightOffset = toAppend.offset, i;
	for (i = 0; i < toAppend.length; i++) {
		array[leftOffset + i] = toAppend.array[rightOffset + i];
	}

	var newSlice = new slice.constructor(array);
	newSlice.offset = slice.offset;
	newSlice.length = newLength;
	newSlice.capacity = slice.capacity;
	return newSlice;
};

var go$panic = function(value) {
	var message;
	if (value.constructor === Go$String) {
		message = value.go$val;
	} else if (value.Error !== undefined) {
		message = value.Error();
	} else if (value.String !== undefined) {
		message = value.String();
	} else {
		message = value;
	}
	var err = new Error(message);
	err.go$panicValue = value;
	return err;
};
var go$notSupported = function(feature) {
	var err = new Error("not supported by GopherJS: " + feature);
	err.go$notSupported = feature;
	throw err;
};
var go$throwRuntimeError; // set by package "runtime"

var go$errorStack = [], go$jsErr = null;

var go$pushErr = function(err) {
	if (err.go$panicValue === undefined) {
		if (err.go$exit || err.go$notSupported) {
			go$jsErr = err;
			return;
		}
		err.go$panicValue = new go$packages["github.com/gopherjs/gopherjs/js"].Error.Ptr(err);
	}
	go$errorStack.push({ frame: go$getStackDepth(), error: err });
};

var go$callDeferred = function(deferred) {
	if (go$jsErr !== null) {
		throw go$jsErr;
	}
	var i;
	for (i = deferred.length - 1; i >= 0; i--) {
		var call = deferred[i];
		try {
			if (call.recv !== undefined) {
				call.recv[call.method].apply(call.recv, call.args);
				continue;
			}
			call.fun.apply(undefined, call.args);
		} catch (err) {
			go$errorStack.push({ frame: go$getStackDepth(), error: err });
		}
	}
	var err = go$errorStack[go$errorStack.length - 1];
	if (err !== undefined && err.frame === go$getStackDepth()) {
		go$errorStack.pop();
		throw err.error;
	}
};

var go$recover = function() {
	var err = go$errorStack[go$errorStack.length - 1];
	if (err === undefined || err.frame !== go$getStackDepth()) {
		return null;
	}
	go$errorStack.pop();
	return err.error.go$panicValue;
};

var go$getStack = function() {
	return (new Error()).stack.split("\n");
};

var go$getStackDepth = function() {
	var s = go$getStack(), d = 0, i;
	for (i = 0; i < s.length; i++) {
		if (s[i].indexOf("go$") === -1) {
			d++;
		}
	}
	return d;
};

var go$interfaceIsEqual = function(a, b) {
	if (a === b) {
		return true;
	}
	if (a === null || b === null || a === undefined || b === undefined || a.constructor !== b.constructor) {
		return false;
	}
	switch (a.constructor.kind) {
	case "Float32":
		return go$float32IsEqual(a.go$val, b.go$val);
	case "Complex64":
		return go$float32IsEqual(a.go$val.real, b.go$val.real) && go$float32IsEqual(a.go$val.imag, b.go$val.imag);
	case "Complex128":
		return a.go$val.real === b.go$val.real && a.go$val.imag === b.go$val.imag;
	case "Int64":
	case "Uint64":
		return a.go$val.high === b.go$val.high && a.go$val.low === b.go$val.low;
	case "Array":
		return go$arrayIsEqual(a.go$val, b.go$val);
	case "Ptr":
		if (a.constructor.Struct) {
			return false;
		}
		return go$pointerIsEqual(a, b);
	case "Func":
	case "Map":
	case "Slice":
	case "Struct":
		go$throwRuntimeError("comparing uncomparable type " + a.constructor);
	case undefined: // js.Object
		return false;
	default:
		return a.go$val === b.go$val;
	}
};
var go$float32IsEqual = function(a, b) {
	if (a === b) {
		return true;
	}
	if (a === 0 || b === 0 || a === 1/0 || b === 1/0 || a === -1/0 || b === -1/0 || a !== a || b !== b) {
		return false;
	}
	var math = go$packages["math"];
	return math !== undefined && math.Float32bits(a) === math.Float32bits(b);
}
var go$arrayIsEqual = function(a, b) {
	if (a.length != b.length) {
		return false;
	}
	var i;
	for (i = 0; i < a.length; i++) {
		if (a[i] !== b[i]) {
			return false;
		}
	}
	return true;
};
var go$sliceIsEqual = function(a, ai, b, bi) {
	return a.array === b.array && a.offset + ai === b.offset + bi;
};
var go$pointerIsEqual = function(a, b) {
	if (a === b) {
		return true;
	}
	if (a.go$get === go$throwNilPointerError || b.go$get === go$throwNilPointerError) {
		return a.go$get === go$throwNilPointerError && b.go$get === go$throwNilPointerError;
	}
	var old = a.go$get();
	var dummy = new Object();
	a.go$set(dummy);
	var equal = b.go$get() === dummy;
	a.go$set(old);
	return equal;
};

var go$typeAssertionFailed = function(obj, expected) {
	var got = "";
	if (obj !== null) {
		got = obj.constructor.string;
	}
	throw go$panic(new go$packages["runtime"].TypeAssertionError.Ptr("", got, expected.string, ""));
};

var go$now = function() { var msec = (new Date()).getTime(); return [new Go$Int64(0, Math.floor(msec / 1000)), (msec % 1000) * 1000000]; };

var go$packages = {};
go$packages["github.com/gopherjs/gopherjs/js"] = (function() {
	var go$pkg = {}, Object, Error;
	Object = go$pkg.Object = go$newType(8, "Interface", "js.Object", "Object", "github.com/gopherjs/gopherjs/js", null);
	Error = go$pkg.Error = go$newType(0, "Struct", "js.Error", "Error", "github.com/gopherjs/gopherjs/js", function(Object_) {
		this.go$val = this;
		this.Object = Object_ !== undefined ? Object_ : null;
	});
	Error.Ptr.prototype.Error = function() {
		var err;
		err = this;
		return "JavaScript error: " + go$internalize(err.Object.message, Go$String);
	};
	Error.prototype.Error = function() { return this.go$val.Error(); };
	go$pkg.init = function() {
		Object.init([["Bool", "", (go$funcType([], [Go$Bool], false))], ["Call", "", (go$funcType([Go$String, (go$sliceType(go$emptyInterface))], [Object], true))], ["Delete", "", (go$funcType([Go$String], [], false))], ["Float", "", (go$funcType([], [Go$Float64], false))], ["Get", "", (go$funcType([Go$String], [Object], false))], ["Index", "", (go$funcType([Go$Int], [Object], false))], ["Int", "", (go$funcType([], [Go$Int], false))], ["Int64", "", (go$funcType([], [Go$Int64], false))], ["Interface", "", (go$funcType([], [go$emptyInterface], false))], ["Invoke", "", (go$funcType([(go$sliceType(go$emptyInterface))], [Object], true))], ["IsNull", "", (go$funcType([], [Go$Bool], false))], ["IsUndefined", "", (go$funcType([], [Go$Bool], false))], ["Length", "", (go$funcType([], [Go$Int], false))], ["New", "", (go$funcType([(go$sliceType(go$emptyInterface))], [Object], true))], ["Set", "", (go$funcType([Go$String, go$emptyInterface], [], false))], ["SetIndex", "", (go$funcType([Go$Int, go$emptyInterface], [], false))], ["Str", "", (go$funcType([], [Go$String], false))], ["Uint64", "", (go$funcType([], [Go$Uint64], false))], ["Unsafe", "", (go$funcType([], [Go$Uintptr], false))]]);
		Error.methods = [["Bool", "", [], [Go$Bool], false, 0], ["Call", "", [Go$String, (go$sliceType(go$emptyInterface))], [Object], true, 0], ["Delete", "", [Go$String], [], false, 0], ["Float", "", [], [Go$Float64], false, 0], ["Get", "", [Go$String], [Object], false, 0], ["Index", "", [Go$Int], [Object], false, 0], ["Int", "", [], [Go$Int], false, 0], ["Int64", "", [], [Go$Int64], false, 0], ["Interface", "", [], [go$emptyInterface], false, 0], ["Invoke", "", [(go$sliceType(go$emptyInterface))], [Object], true, 0], ["IsNull", "", [], [Go$Bool], false, 0], ["IsUndefined", "", [], [Go$Bool], false, 0], ["Length", "", [], [Go$Int], false, 0], ["New", "", [(go$sliceType(go$emptyInterface))], [Object], true, 0], ["Set", "", [Go$String, go$emptyInterface], [], false, 0], ["SetIndex", "", [Go$Int, go$emptyInterface], [], false, 0], ["Str", "", [], [Go$String], false, 0], ["Uint64", "", [], [Go$Uint64], false, 0], ["Unsafe", "", [], [Go$Uintptr], false, 0]];
		(go$ptrType(Error)).methods = [["Bool", "", [], [Go$Bool], false, 0], ["Call", "", [Go$String, (go$sliceType(go$emptyInterface))], [Object], true, 0], ["Delete", "", [Go$String], [], false, 0], ["Error", "", [], [Go$String], false, -1], ["Float", "", [], [Go$Float64], false, 0], ["Get", "", [Go$String], [Object], false, 0], ["Index", "", [Go$Int], [Object], false, 0], ["Int", "", [], [Go$Int], false, 0], ["Int64", "", [], [Go$Int64], false, 0], ["Interface", "", [], [go$emptyInterface], false, 0], ["Invoke", "", [(go$sliceType(go$emptyInterface))], [Object], true, 0], ["IsNull", "", [], [Go$Bool], false, 0], ["IsUndefined", "", [], [Go$Bool], false, 0], ["Length", "", [], [Go$Int], false, 0], ["New", "", [(go$sliceType(go$emptyInterface))], [Object], true, 0], ["Set", "", [Go$String, go$emptyInterface], [], false, 0], ["SetIndex", "", [Go$Int, go$emptyInterface], [], false, 0], ["Str", "", [], [Go$String], false, 0], ["Uint64", "", [], [Go$Uint64], false, 0], ["Unsafe", "", [], [Go$Uintptr], false, 0]];
		Error.init([["Object", "", "", Object, ""]]);
		var e;
		e = new Error.Ptr(null);
	}
	return go$pkg;
})();
go$packages["runtime"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], TypeAssertionError, errorString, getgoroot, SetFinalizer, GOROOT, goexit, sizeof_C_MStats;
	TypeAssertionError = go$pkg.TypeAssertionError = go$newType(0, "Struct", "runtime.TypeAssertionError", "TypeAssertionError", "runtime", function(interfaceString_, concreteString_, assertedString_, missingMethod_) {
		this.go$val = this;
		this.interfaceString = interfaceString_ !== undefined ? interfaceString_ : "";
		this.concreteString = concreteString_ !== undefined ? concreteString_ : "";
		this.assertedString = assertedString_ !== undefined ? assertedString_ : "";
		this.missingMethod = missingMethod_ !== undefined ? missingMethod_ : "";
	});
	errorString = go$pkg.errorString = go$newType(8, "String", "runtime.errorString", "errorString", "runtime", null);
	getgoroot = function() {
		var process, goroot;
		process = go$global.process;
		if (process === undefined) {
			return "/";
		}
		goroot = process.env.GOROOT;
		if (goroot === undefined) {
			return "";
		}
		return go$internalize(goroot, Go$String);
	};
	SetFinalizer = go$pkg.SetFinalizer = function(x, f) {
	};
	TypeAssertionError.Ptr.prototype.RuntimeError = function() {
	};
	TypeAssertionError.prototype.RuntimeError = function() { return this.go$val.RuntimeError(); };
	TypeAssertionError.Ptr.prototype.Error = function() {
		var e, inter;
		e = this;
		inter = e.interfaceString;
		if (inter === "") {
			inter = "interface";
		}
		if (e.concreteString === "") {
			return "interface conversion: " + inter + " is nil, not " + e.assertedString;
		}
		if (e.missingMethod === "") {
			return "interface conversion: " + inter + " is " + e.concreteString + ", not " + e.assertedString;
		}
		return "interface conversion: " + e.concreteString + " is not " + e.assertedString + ": missing method " + e.missingMethod;
	};
	TypeAssertionError.prototype.Error = function() { return this.go$val.Error(); };
	errorString.prototype.RuntimeError = function() {
		var e;
		e = this.go$val;
	};
	go$ptrType(errorString).prototype.RuntimeError = function() { return new errorString(this.go$get()).RuntimeError(); };
	errorString.prototype.Error = function() {
		var e;
		e = this.go$val;
		return "runtime error: " + e;
	};
	go$ptrType(errorString).prototype.Error = function() { return new errorString(this.go$get()).Error(); };
	GOROOT = go$pkg.GOROOT = function() {
		var s;
		s = getgoroot();
		if (!(s === "")) {
			return s;
		}
		return "/usr/local/Cellar/go/1.2/libexec";
	};
	go$pkg.init = function() {
		(go$ptrType(TypeAssertionError)).methods = [["Error", "", [], [Go$String], false, -1], ["RuntimeError", "", [], [], false, -1]];
		TypeAssertionError.init([["interfaceString", "interfaceString", "runtime", Go$String, ""], ["concreteString", "concreteString", "runtime", Go$String, ""], ["assertedString", "assertedString", "runtime", Go$String, ""], ["missingMethod", "missingMethod", "runtime", Go$String, ""]]);
		errorString.methods = [["Error", "", [], [Go$String], false, -1], ["RuntimeError", "", [], [], false, -1]];
		(go$ptrType(errorString)).methods = [["Error", "", [], [Go$String], false, -1], ["RuntimeError", "", [], [], false, -1]];
		sizeof_C_MStats = 3712;
		goexit = go$global.eval(go$externalize("(function() {\n\tvar err = new Error();\n\terr.go$exit = true;\n\tthrow err;\n})", Go$String));
		var e;
		go$throwRuntimeError = go$externalize((function(msg) {
			throw go$panic(new errorString(msg));
		}), (go$funcType([Go$String], [], false)));
		e = new TypeAssertionError.Ptr("", "", "", "");
		if (!((sizeof_C_MStats === 3712))) {
			console.log(sizeof_C_MStats, 3712);
			throw go$panic(new Go$String("MStats vs MemStatsType size mismatch"));
		}
	}
	return go$pkg;
})();
go$packages["github.com/gopherjs/jquery"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], JQuery, Event, JQueryCoordinates, NewJQuery, ParseHTML;
	JQuery = go$pkg.JQuery = go$newType(0, "Struct", "jquery.JQuery", "JQuery", "github.com/gopherjs/jquery", function(o_, Jquery_, Selector_, Length_, Context_) {
		this.go$val = this;
		this.o = o_ !== undefined ? o_ : null;
		this.Jquery = Jquery_ !== undefined ? Jquery_ : "";
		this.Selector = Selector_ !== undefined ? Selector_ : "";
		this.Length = Length_ !== undefined ? Length_ : "";
		this.Context = Context_ !== undefined ? Context_ : "";
	});
	Event = go$pkg.Event = go$newType(0, "Struct", "jquery.Event", "Event", "github.com/gopherjs/jquery", function(Object_, KeyCode_, Target_, CurrentTarget_, DelegateTarget_, RelatedTarget_, Data_, Result_, Which_, Namespace_, MetaKey_, PageX_, PageY_, Type_) {
		this.go$val = this;
		this.Object = Object_ !== undefined ? Object_ : null;
		this.KeyCode = KeyCode_ !== undefined ? KeyCode_ : 0;
		this.Target = Target_ !== undefined ? Target_ : null;
		this.CurrentTarget = CurrentTarget_ !== undefined ? CurrentTarget_ : null;
		this.DelegateTarget = DelegateTarget_ !== undefined ? DelegateTarget_ : null;
		this.RelatedTarget = RelatedTarget_ !== undefined ? RelatedTarget_ : null;
		this.Data = Data_ !== undefined ? Data_ : null;
		this.Result = Result_ !== undefined ? Result_ : null;
		this.Which = Which_ !== undefined ? Which_ : 0;
		this.Namespace = Namespace_ !== undefined ? Namespace_ : "";
		this.MetaKey = MetaKey_ !== undefined ? MetaKey_ : false;
		this.PageX = PageX_ !== undefined ? PageX_ : 0;
		this.PageY = PageY_ !== undefined ? PageY_ : 0;
		this.Type = Type_ !== undefined ? Type_ : "";
	});
	JQueryCoordinates = go$pkg.JQueryCoordinates = go$newType(0, "Struct", "jquery.JQueryCoordinates", "JQueryCoordinates", "github.com/gopherjs/jquery", function(Left_, Top_) {
		this.go$val = this;
		this.Left = Left_ !== undefined ? Left_ : 0;
		this.Top = Top_ !== undefined ? Top_ : 0;
	});
	Event.Ptr.prototype.PreventDefault = function() {
		var event;
		event = this;
		event.Object.preventDefault();
	};
	Event.prototype.PreventDefault = function() { return this.go$val.PreventDefault(); };
	Event.Ptr.prototype.IsDefaultPrevented = function() {
		var event;
		event = this;
		return !!(event.Object.isDefaultPrevented());
	};
	Event.prototype.IsDefaultPrevented = function() { return this.go$val.IsDefaultPrevented(); };
	Event.Ptr.prototype.IsImmediatePropogationStopped = function() {
		var event;
		event = this;
		return !!(event.Object.isImmediatePropogationStopped());
	};
	Event.prototype.IsImmediatePropogationStopped = function() { return this.go$val.IsImmediatePropogationStopped(); };
	Event.Ptr.prototype.IsPropagationStopped = function() {
		var event;
		event = this;
		return !!(event.Object.isPropagationStopped());
	};
	Event.prototype.IsPropagationStopped = function() { return this.go$val.IsPropagationStopped(); };
	Event.Ptr.prototype.StopImmediatePropagation = function() {
		var event;
		event = this;
		event.Object.stopImmediatePropagation();
	};
	Event.prototype.StopImmediatePropagation = function() { return this.go$val.StopImmediatePropagation(); };
	Event.Ptr.prototype.StopPropagation = function() {
		var event;
		event = this;
		event.Object.stopPropagation();
	};
	Event.prototype.StopPropagation = function() { return this.go$val.StopPropagation(); };
	NewJQuery = go$pkg.NewJQuery = function(args) {
		return new JQuery.Ptr(new (go$global.Function.prototype.bind.apply(go$global.jQuery, [undefined].concat(go$externalize(args, (go$sliceType(go$emptyInterface)))))), "", "", "", "");
	};
	ParseHTML = go$pkg.ParseHTML = function(text) {
		var x;
		return (x = go$internalize(go$global.jQuery.parseHTML(go$externalize(text, Go$String)), go$emptyInterface), (x !== null && x.constructor === (go$sliceType(go$emptyInterface)) ? x.go$val : go$typeAssertionFailed(x, (go$sliceType(go$emptyInterface)))));
	};
	JQuery.Ptr.prototype.Each = function(fn) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.each(go$externalize(fn, (go$funcType([Go$Int, go$emptyInterface], [go$emptyInterface], false))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Each = function(fn) { return this.go$val.Each(fn); };
	JQuery.Ptr.prototype.Underlying = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return j.o;
	};
	JQuery.prototype.Underlying = function() { return this.go$val.Underlying(); };
	JQuery.Ptr.prototype.Get = function(i) {
		var _struct, j, obj;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (obj = j.o, obj.get.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
	};
	JQuery.prototype.Get = function(i) { return this.go$val.Get(i); };
	JQuery.Ptr.prototype.Append = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom2args("append", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Append = function(i) { return this.go$val.Append(i); };
	JQuery.Ptr.prototype.Empty = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.empty();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Empty = function() { return this.go$val.Empty(); };
	JQuery.Ptr.prototype.Detach = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.detach.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Detach = function(i) { return this.go$val.Detach(i); };
	JQuery.Ptr.prototype.Eq = function(idx) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.eq(idx);
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Eq = function(idx) { return this.go$val.Eq(idx); };
	JQuery.Ptr.prototype.FadeIn = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.fadeIn.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.FadeIn = function(i) { return this.go$val.FadeIn(i); };
	JQuery.Ptr.prototype.Delay = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.delay.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Delay = function(i) { return this.go$val.Delay(i); };
	JQuery.Ptr.prototype.ToArray = function() {
		var _struct, j, x;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (x = go$internalize(j.o.toArray(), go$emptyInterface), (x !== null && x.constructor === (go$sliceType(go$emptyInterface)) ? x.go$val : go$typeAssertionFailed(x, (go$sliceType(go$emptyInterface)))));
	};
	JQuery.prototype.ToArray = function() { return this.go$val.ToArray(); };
	JQuery.Ptr.prototype.Remove = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.remove.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Remove = function(i) { return this.go$val.Remove(i); };
	JQuery.Ptr.prototype.Stop = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.stop.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Stop = function(i) { return this.go$val.Stop(i); };
	JQuery.Ptr.prototype.AddBack = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.addBack.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.AddBack = function(i) { return this.go$val.AddBack(i); };
	JQuery.Ptr.prototype.Css = function(name) {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$internalize(j.o.css(go$externalize(name, Go$String)), Go$String);
	};
	JQuery.prototype.Css = function(name) { return this.go$val.Css(name); };
	JQuery.Ptr.prototype.CssArray = function(arr) {
		var _struct, j, x;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (x = go$internalize(j.o.css(go$externalize(arr, (go$sliceType(Go$String)))), go$emptyInterface), (x !== null && x.constructor === (go$mapType(Go$String, go$emptyInterface)) ? x.go$val : go$typeAssertionFailed(x, (go$mapType(Go$String, go$emptyInterface)))));
	};
	JQuery.prototype.CssArray = function(arr) { return this.go$val.CssArray(arr); };
	JQuery.Ptr.prototype.SetCss = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.css.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetCss = function(i) { return this.go$val.SetCss(i); };
	JQuery.Ptr.prototype.Text = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$internalize(j.o.text(), Go$String);
	};
	JQuery.prototype.Text = function() { return this.go$val.Text(); };
	JQuery.Ptr.prototype.SetText = function(i) {
		var _struct, j, _ref, _type, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = i;
		_type = _ref !== null ? _ref.constructor : null;
		if (_type === (go$funcType([Go$Int, Go$String], [Go$String], false)) || _type === Go$String) {
		} else {
			console.log("SetText Argument should be 'string' or 'func(int, string) string'");
		}
		j.o = j.o.text(go$externalize(i, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetText = function(i) { return this.go$val.SetText(i); };
	JQuery.Ptr.prototype.Val = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$internalize(j.o.val(), Go$String);
	};
	JQuery.prototype.Val = function() { return this.go$val.Val(); };
	JQuery.Ptr.prototype.SetVal = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o.val(go$externalize(i, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetVal = function(i) { return this.go$val.SetVal(i); };
	JQuery.Ptr.prototype.Prop = function(property) {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$internalize(j.o.prop(go$externalize(property, Go$String)), go$emptyInterface);
	};
	JQuery.prototype.Prop = function(property) { return this.go$val.Prop(property); };
	JQuery.Ptr.prototype.SetProp = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.prop.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetProp = function(i) { return this.go$val.SetProp(i); };
	JQuery.Ptr.prototype.RemoveProp = function(property) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.removeProp(go$externalize(property, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.RemoveProp = function(property) { return this.go$val.RemoveProp(property); };
	JQuery.Ptr.prototype.Attr = function(property) {
		var _struct, j, attr;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		attr = j.o.attr(go$externalize(property, Go$String));
		if (attr === undefined) {
			return "";
		}
		return go$internalize(attr, Go$String);
	};
	JQuery.prototype.Attr = function(property) { return this.go$val.Attr(property); };
	JQuery.Ptr.prototype.SetAttr = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.attr.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetAttr = function(i) { return this.go$val.SetAttr(i); };
	JQuery.Ptr.prototype.RemoveAttr = function(property) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.removeAttr(go$externalize(property, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.RemoveAttr = function(property) { return this.go$val.RemoveAttr(property); };
	JQuery.Ptr.prototype.HasClass = function(class$1) {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return !!(j.o.hasClass(go$externalize(class$1, Go$String)));
	};
	JQuery.prototype.HasClass = function(class$1) { return this.go$val.HasClass(class$1); };
	JQuery.Ptr.prototype.AddClass = function(i) {
		var _struct, j, _ref, _type, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = i;
		_type = _ref !== null ? _ref.constructor : null;
		if (_type === (go$funcType([Go$Int, Go$String], [Go$String], false)) || _type === Go$String) {
		} else {
			console.log("addClass Argument should be 'string' or 'func(int, string) string'");
		}
		j.o = j.o.addClass(go$externalize(i, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.AddClass = function(i) { return this.go$val.AddClass(i); };
	JQuery.Ptr.prototype.RemoveClass = function(property) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.removeClass(go$externalize(property, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.RemoveClass = function(property) { return this.go$val.RemoveClass(property); };
	JQuery.Ptr.prototype.ToggleClass = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.toggleClass.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.ToggleClass = function(i) { return this.go$val.ToggleClass(i); };
	JQuery.Ptr.prototype.Focus = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.focus();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Focus = function() { return this.go$val.Focus(); };
	JQuery.Ptr.prototype.Blur = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.blur();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Blur = function() { return this.go$val.Blur(); };
	JQuery.Ptr.prototype.ReplaceAll = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("replaceAll", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.ReplaceAll = function(i) { return this.go$val.ReplaceAll(i); };
	JQuery.Ptr.prototype.ReplaceWith = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("replaceWith", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.ReplaceWith = function(i) { return this.go$val.ReplaceWith(i); };
	JQuery.Ptr.prototype.After = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom2args("after", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.After = function(i) { return this.go$val.After(i); };
	JQuery.Ptr.prototype.Before = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom2args("before", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Before = function(i) { return this.go$val.Before(i); };
	JQuery.Ptr.prototype.Prepend = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom2args("prepend", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Prepend = function(i) { return this.go$val.Prepend(i); };
	JQuery.Ptr.prototype.PrependTo = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("prependTo", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.PrependTo = function(i) { return this.go$val.PrependTo(i); };
	JQuery.Ptr.prototype.AppendTo = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("appendTo", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.AppendTo = function(i) { return this.go$val.AppendTo(i); };
	JQuery.Ptr.prototype.InsertAfter = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("insertAfter", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.InsertAfter = function(i) { return this.go$val.InsertAfter(i); };
	JQuery.Ptr.prototype.InsertBefore = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("insertBefore", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.InsertBefore = function(i) { return this.go$val.InsertBefore(i); };
	JQuery.Ptr.prototype.Show = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.show();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Show = function() { return this.go$val.Show(); };
	JQuery.Ptr.prototype.Hide = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o.hide();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Hide = function() { return this.go$val.Hide(); };
	JQuery.Ptr.prototype.Toggle = function(showOrHide) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.toggle(go$externalize(showOrHide, Go$Bool));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Toggle = function(showOrHide) { return this.go$val.Toggle(showOrHide); };
	JQuery.Ptr.prototype.Contents = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.contents();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Contents = function() { return this.go$val.Contents(); };
	JQuery.Ptr.prototype.Html = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$internalize(j.o.html(), Go$String);
	};
	JQuery.prototype.Html = function() { return this.go$val.Html(); };
	JQuery.Ptr.prototype.SetHtml = function(i) {
		var _struct, j, _ref, _type, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = i;
		_type = _ref !== null ? _ref.constructor : null;
		if (_type === (go$funcType([Go$Int, Go$String], [Go$String], false)) || _type === Go$String) {
		} else {
			console.log("SetHtml Argument should be 'string' or 'func(int, string) string'");
		}
		j.o = j.o.html(go$externalize(i, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetHtml = function(i) { return this.go$val.SetHtml(i); };
	JQuery.Ptr.prototype.Closest = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom2args("closest", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Closest = function(i) { return this.go$val.Closest(i); };
	JQuery.Ptr.prototype.End = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.end();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.End = function() { return this.go$val.End(); };
	JQuery.Ptr.prototype.Add = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom2args("add", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Add = function(i) { return this.go$val.Add(i); };
	JQuery.Ptr.prototype.Clone = function(b) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.clone.apply(obj, go$externalize(b, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Clone = function(b) { return this.go$val.Clone(b); };
	JQuery.Ptr.prototype.Height = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$parseInt(j.o.height()) >> 0;
	};
	JQuery.prototype.Height = function() { return this.go$val.Height(); };
	JQuery.Ptr.prototype.SetHeight = function(value) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.height(go$externalize(value, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetHeight = function(value) { return this.go$val.SetHeight(value); };
	JQuery.Ptr.prototype.Width = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$parseInt(j.o.width()) >> 0;
	};
	JQuery.prototype.Width = function() { return this.go$val.Width(); };
	JQuery.Ptr.prototype.SetWidth = function(i) {
		var _struct, j, _ref, _type, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = i;
		_type = _ref !== null ? _ref.constructor : null;
		if (_type === (go$funcType([Go$Int, Go$String], [Go$String], false)) || _type === Go$String) {
		} else {
			console.log("SetWidth Argument should be 'string' or 'func(int, string) string'");
		}
		j.o = j.o.width(go$externalize(i, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetWidth = function(i) { return this.go$val.SetWidth(i); };
	JQuery.Ptr.prototype.InnerHeight = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$parseInt(j.o.innerHeight()) >> 0;
	};
	JQuery.prototype.InnerHeight = function() { return this.go$val.InnerHeight(); };
	JQuery.Ptr.prototype.InnerWidth = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$parseInt(j.o.innerWidth()) >> 0;
	};
	JQuery.prototype.InnerWidth = function() { return this.go$val.InnerWidth(); };
	JQuery.Ptr.prototype.Offset = function() {
		var _struct, j, obj;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		obj = j.o.offset();
		return new JQueryCoordinates.Ptr(go$parseInt(obj.left) >> 0, go$parseInt(obj.top) >> 0);
	};
	JQuery.prototype.Offset = function() { return this.go$val.Offset(); };
	JQuery.Ptr.prototype.SetOffset = function(jc) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.offset(go$externalize(jc, JQueryCoordinates));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetOffset = function(jc) { return this.go$val.SetOffset(jc); };
	JQuery.Ptr.prototype.OuterHeight = function(includeMargin) {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		if (includeMargin.length === 0) {
			return go$parseInt(j.o.outerHeight()) >> 0;
		}
		return go$parseInt(j.o.outerHeight(go$externalize(((0 < 0 || 0 >= includeMargin.length) ? go$throwRuntimeError("index out of range") : includeMargin.array[includeMargin.offset + 0]), Go$Bool))) >> 0;
	};
	JQuery.prototype.OuterHeight = function(includeMargin) { return this.go$val.OuterHeight(includeMargin); };
	JQuery.Ptr.prototype.OuterWidth = function(includeMargin) {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		if (includeMargin.length === 0) {
			return go$parseInt(j.o.outerWidth()) >> 0;
		}
		return go$parseInt(j.o.outerWidth(go$externalize(((0 < 0 || 0 >= includeMargin.length) ? go$throwRuntimeError("index out of range") : includeMargin.array[includeMargin.offset + 0]), Go$Bool))) >> 0;
	};
	JQuery.prototype.OuterWidth = function(includeMargin) { return this.go$val.OuterWidth(includeMargin); };
	JQuery.Ptr.prototype.Position = function() {
		var _struct, j, obj;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		obj = j.o.position();
		return new JQueryCoordinates.Ptr(go$parseInt(obj.left) >> 0, go$parseInt(obj.top) >> 0);
	};
	JQuery.prototype.Position = function() { return this.go$val.Position(); };
	JQuery.Ptr.prototype.ScrollLeft = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$parseInt(j.o.scrollLeft()) >> 0;
	};
	JQuery.prototype.ScrollLeft = function() { return this.go$val.ScrollLeft(); };
	JQuery.Ptr.prototype.SetScrollLeft = function(value) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.scrollLeft(value);
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetScrollLeft = function(value) { return this.go$val.SetScrollLeft(value); };
	JQuery.Ptr.prototype.ScrollTop = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$parseInt(j.o.scrollTop()) >> 0;
	};
	JQuery.prototype.ScrollTop = function() { return this.go$val.ScrollTop(); };
	JQuery.Ptr.prototype.SetScrollTop = function(value) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.scrollTop(value);
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetScrollTop = function(value) { return this.go$val.SetScrollTop(value); };
	JQuery.Ptr.prototype.ClearQueue = function(queueName) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.clearQueue(go$externalize(queueName, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.ClearQueue = function(queueName) { return this.go$val.ClearQueue(queueName); };
	JQuery.Ptr.prototype.SetData = function(key, value) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.data(go$externalize(key, Go$String), go$externalize(value, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.SetData = function(key, value) { return this.go$val.SetData(key, value); };
	JQuery.Ptr.prototype.Data = function(key) {
		var _struct, j, result;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		result = j.o.data(go$externalize(key, Go$String));
		if (result === undefined) {
			return null;
		}
		return go$internalize(result, go$emptyInterface);
	};
	JQuery.prototype.Data = function(key) { return this.go$val.Data(key); };
	JQuery.Ptr.prototype.Dequeue = function(queueName) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.dequeue(go$externalize(queueName, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Dequeue = function(queueName) { return this.go$val.Dequeue(queueName); };
	JQuery.Ptr.prototype.RemoveData = function(name) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.removeData(go$externalize(name, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.RemoveData = function(name) { return this.go$val.RemoveData(name); };
	JQuery.Ptr.prototype.OffsetParent = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.offsetParent();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.OffsetParent = function() { return this.go$val.OffsetParent(); };
	JQuery.Ptr.prototype.Parent = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.parent.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Parent = function(i) { return this.go$val.Parent(i); };
	JQuery.Ptr.prototype.Parents = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.parents.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Parents = function(i) { return this.go$val.Parents(i); };
	JQuery.Ptr.prototype.ParentsUntil = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.parentsUntil.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.ParentsUntil = function(i) { return this.go$val.ParentsUntil(i); };
	JQuery.Ptr.prototype.Prev = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.prev.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Prev = function(i) { return this.go$val.Prev(i); };
	JQuery.Ptr.prototype.PrevAll = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.prevAll.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.PrevAll = function(i) { return this.go$val.PrevAll(i); };
	JQuery.Ptr.prototype.PrevUntil = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.prevUntil.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.PrevUntil = function(i) { return this.go$val.PrevUntil(i); };
	JQuery.Ptr.prototype.Siblings = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.siblings.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Siblings = function(i) { return this.go$val.Siblings(i); };
	JQuery.Ptr.prototype.Slice = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.slice.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Slice = function(i) { return this.go$val.Slice(i); };
	JQuery.Ptr.prototype.Children = function(selector) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.children(go$externalize(selector, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Children = function(selector) { return this.go$val.Children(selector); };
	JQuery.Ptr.prototype.Unwrap = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.unwrap();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Unwrap = function() { return this.go$val.Unwrap(); };
	JQuery.Ptr.prototype.Wrap = function(obj) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.wrap(go$externalize(obj, go$emptyInterface));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Wrap = function(obj) { return this.go$val.Wrap(obj); };
	JQuery.Ptr.prototype.WrapAll = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("wrapAll", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.WrapAll = function(i) { return this.go$val.WrapAll(i); };
	JQuery.Ptr.prototype.WrapInner = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.dom1arg("wrapInner", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.WrapInner = function(i) { return this.go$val.WrapInner(i); };
	JQuery.Ptr.prototype.Next = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.next.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Next = function(i) { return this.go$val.Next(i); };
	JQuery.Ptr.prototype.NextAll = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.nextAll.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.NextAll = function(i) { return this.go$val.NextAll(i); };
	JQuery.Ptr.prototype.NextUntil = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.nextUntil.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.NextUntil = function(i) { return this.go$val.NextUntil(i); };
	JQuery.Ptr.prototype.Not = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.not.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Not = function(i) { return this.go$val.Not(i); };
	JQuery.Ptr.prototype.Filter = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.filter.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Filter = function(i) { return this.go$val.Filter(i); };
	JQuery.Ptr.prototype.Find = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.find.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Find = function(i) { return this.go$val.Find(i); };
	JQuery.Ptr.prototype.First = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.first();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.First = function() { return this.go$val.First(); };
	JQuery.Ptr.prototype.Has = function(selector) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.has(go$externalize(selector, Go$String));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Has = function(selector) { return this.go$val.Has(selector); };
	JQuery.Ptr.prototype.Is = function(i) {
		var _struct, j, obj;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return !!((obj = j.o, obj.is.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface))))));
	};
	JQuery.prototype.Is = function(i) { return this.go$val.Is(i); };
	JQuery.Ptr.prototype.Last = function() {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.last();
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Last = function() { return this.go$val.Last(); };
	JQuery.Ptr.prototype.Ready = function(handler) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = j.o.ready(go$externalize(handler, (go$funcType([], [], false))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Ready = function(handler) { return this.go$val.Ready(handler); };
	JQuery.Ptr.prototype.Resize = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.resize.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Resize = function(i) { return this.go$val.Resize(i); };
	JQuery.Ptr.prototype.Scroll = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.handleEvent("scroll", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Scroll = function(i) { return this.go$val.Scroll(i); };
	JQuery.Ptr.prototype.FadeOut = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.fadeOut.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.FadeOut = function(i) { return this.go$val.FadeOut(i); };
	JQuery.Ptr.prototype.Select = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.handleEvent("select", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Select = function(i) { return this.go$val.Select(i); };
	JQuery.Ptr.prototype.Submit = function(i) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.handleEvent("submit", i), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Submit = function(i) { return this.go$val.Submit(i); };
	JQuery.Ptr.prototype.handleEvent = function(evt, i) {
		var _struct, j, _ref, x, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = i.length;
		if (_ref === 0) {
			j.o = j.o[go$externalize(evt, Go$String)]();
		} else if (_ref === 1) {
			j.o = j.o[go$externalize(evt, Go$String)](go$externalize((function(e) {
				var x;
				(x = ((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), (x !== null && x.constructor === (go$funcType([Event], [], false)) ? x.go$val : go$typeAssertionFailed(x, (go$funcType([Event], [], false)))))(new Event.Ptr(e, 0, null, null, null, null, null, null, 0, "", false, 0, 0, ""));
			}), (go$funcType([js.Object], [], false))));
		} else if (_ref === 2) {
			j.o = j.o[go$externalize(evt, Go$String)](go$externalize((x = ((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), (x !== null && x.constructor === (go$mapType(Go$String, go$emptyInterface)) ? x.go$val : go$typeAssertionFailed(x, (go$mapType(Go$String, go$emptyInterface))))), (go$mapType(Go$String, go$emptyInterface))), go$externalize((function(e) {
				var x$1;
				(x$1 = ((1 < 0 || 1 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 1]), (x$1 !== null && x$1.constructor === (go$funcType([Event], [], false)) ? x$1.go$val : go$typeAssertionFailed(x$1, (go$funcType([Event], [], false)))))(new Event.Ptr(e, 0, null, null, null, null, null, null, 0, "", false, 0, 0, ""));
			}), (go$funcType([js.Object], [], false))));
		} else {
			console.log(evt + " event expects 0 to 2 arguments");
		}
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.handleEvent = function(evt, i) { return this.go$val.handleEvent(evt, i); };
	JQuery.Ptr.prototype.Trigger = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.trigger.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Trigger = function(i) { return this.go$val.Trigger(i); };
	JQuery.Ptr.prototype.On = function(p) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.events("on", p), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.On = function(p) { return this.go$val.On(p); };
	JQuery.Ptr.prototype.One = function(p) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.events("one", p), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.One = function(p) { return this.go$val.One(p); };
	JQuery.Ptr.prototype.Off = function(p) {
		var _struct, j, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return (_struct$1 = j.events("off", p), new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Off = function(p) { return this.go$val.Off(p); };
	JQuery.Ptr.prototype.events = function(evt, p) {
		var _struct, j, count, isEventFunc, _ref, _type, x, _ref$1, _struct$1, _struct$2, _struct$3, _struct$4, _struct$5, _struct$6, _struct$7, _struct$8, obj, _struct$9;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		count = p.length;
		isEventFunc = false;
		_ref = (x = p.length - 1 >> 0, ((x < 0 || x >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + x]));
		_type = _ref !== null ? _ref.constructor : null;
		if (_type === (go$funcType([Event], [], false))) {
			isEventFunc = true;
		} else {
			isEventFunc = false;
		}
		_ref$1 = count;
		if (_ref$1 === 0) {
			j.o = j.o[go$externalize(evt, Go$String)]();
			return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
		} else if (_ref$1 === 1) {
			j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface));
			return (_struct$2 = j, new JQuery.Ptr(_struct$2.o, _struct$2.Jquery, _struct$2.Selector, _struct$2.Length, _struct$2.Context));
		} else if (_ref$1 === 2) {
			if (isEventFunc) {
				j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface), go$externalize((function(e) {
					var x$1;
					(x$1 = ((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]), (x$1 !== null && x$1.constructor === (go$funcType([Event], [], false)) ? x$1.go$val : go$typeAssertionFailed(x$1, (go$funcType([Event], [], false)))))(new Event.Ptr(e, 0, null, null, null, null, null, null, 0, "", false, 0, 0, ""));
				}), (go$funcType([js.Object], [], false))));
				return (_struct$3 = j, new JQuery.Ptr(_struct$3.o, _struct$3.Jquery, _struct$3.Selector, _struct$3.Length, _struct$3.Context));
			} else {
				j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface), go$externalize(((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]), go$emptyInterface));
				return (_struct$4 = j, new JQuery.Ptr(_struct$4.o, _struct$4.Jquery, _struct$4.Selector, _struct$4.Length, _struct$4.Context));
			}
		} else if (_ref$1 === 3) {
			if (isEventFunc) {
				j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface), go$externalize(((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]), go$emptyInterface), go$externalize((function(e) {
					var x$1;
					(x$1 = ((2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2]), (x$1 !== null && x$1.constructor === (go$funcType([Event], [], false)) ? x$1.go$val : go$typeAssertionFailed(x$1, (go$funcType([Event], [], false)))))(new Event.Ptr(e, 0, null, null, null, null, null, null, 0, "", false, 0, 0, ""));
				}), (go$funcType([js.Object], [], false))));
				return (_struct$5 = j, new JQuery.Ptr(_struct$5.o, _struct$5.Jquery, _struct$5.Selector, _struct$5.Length, _struct$5.Context));
			} else {
				j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface), go$externalize(((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]), go$emptyInterface), go$externalize(((2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2]), go$emptyInterface));
				return (_struct$6 = j, new JQuery.Ptr(_struct$6.o, _struct$6.Jquery, _struct$6.Selector, _struct$6.Length, _struct$6.Context));
			}
		} else if (_ref$1 === 4) {
			if (isEventFunc) {
				j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface), go$externalize(((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]), go$emptyInterface), go$externalize(((2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2]), go$emptyInterface), go$externalize((function(e) {
					var x$1;
					(x$1 = ((3 < 0 || 3 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 3]), (x$1 !== null && x$1.constructor === (go$funcType([Event], [], false)) ? x$1.go$val : go$typeAssertionFailed(x$1, (go$funcType([Event], [], false)))))(new Event.Ptr(e, 0, null, null, null, null, null, null, 0, "", false, 0, 0, ""));
				}), (go$funcType([js.Object], [], false))));
				return (_struct$7 = j, new JQuery.Ptr(_struct$7.o, _struct$7.Jquery, _struct$7.Selector, _struct$7.Length, _struct$7.Context));
			} else {
				j.o = j.o[go$externalize(evt, Go$String)](go$externalize(((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]), go$emptyInterface), go$externalize(((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]), go$emptyInterface), go$externalize(((2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2]), go$emptyInterface), go$externalize(((3 < 0 || 3 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 3]), go$emptyInterface));
				return (_struct$8 = j, new JQuery.Ptr(_struct$8.o, _struct$8.Jquery, _struct$8.Selector, _struct$8.Length, _struct$8.Context));
			}
		} else {
			console.log(evt + " event should no have more than 4 arguments");
			j.o = (obj = j.o, obj[go$externalize(evt, Go$String)].apply(obj, go$externalize(p, (go$sliceType(go$emptyInterface)))));
			return (_struct$9 = j, new JQuery.Ptr(_struct$9.o, _struct$9.Jquery, _struct$9.Selector, _struct$9.Length, _struct$9.Context));
		}
	};
	JQuery.prototype.events = function(evt, p) { return this.go$val.events(evt, p); };
	JQuery.Ptr.prototype.dom2args = function(method, i) {
		var _struct, j, _ref, _tuple, x, _struct$1, selector, selOk, _tuple$1, x$1, _struct$2, context, ctxOk, _struct$3, _struct$4, _struct$5, _struct$6, _tuple$2, x$2, _struct$7, selector$1, selOk$1, _struct$8, _struct$9, _struct$10;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = i.length;
		if (_ref === 2) {
			_tuple = (x = ((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), (x !== null && x.constructor === JQuery ? [x.go$val, true] : [new JQuery.Ptr(), false])); selector = (_struct$1 = _tuple[0], new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context)); selOk = _tuple[1];
			_tuple$1 = (x$1 = ((1 < 0 || 1 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 1]), (x$1 !== null && x$1.constructor === JQuery ? [x$1.go$val, true] : [new JQuery.Ptr(), false])); context = (_struct$2 = _tuple$1[0], new JQuery.Ptr(_struct$2.o, _struct$2.Jquery, _struct$2.Selector, _struct$2.Length, _struct$2.Context)); ctxOk = _tuple$1[1];
			if (!selOk && !ctxOk) {
				j.o = j.o[go$externalize(method, Go$String)](go$externalize(((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), go$emptyInterface), go$externalize(((1 < 0 || 1 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 1]), go$emptyInterface));
				return (_struct$3 = j, new JQuery.Ptr(_struct$3.o, _struct$3.Jquery, _struct$3.Selector, _struct$3.Length, _struct$3.Context));
			} else if (selOk && !ctxOk) {
				j.o = j.o[go$externalize(method, Go$String)](selector.o, go$externalize(((1 < 0 || 1 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 1]), go$emptyInterface));
				return (_struct$4 = j, new JQuery.Ptr(_struct$4.o, _struct$4.Jquery, _struct$4.Selector, _struct$4.Length, _struct$4.Context));
			} else if (!selOk && ctxOk) {
				j.o = j.o[go$externalize(method, Go$String)](go$externalize(((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), go$emptyInterface), context.o);
				return (_struct$5 = j, new JQuery.Ptr(_struct$5.o, _struct$5.Jquery, _struct$5.Selector, _struct$5.Length, _struct$5.Context));
			}
			j.o = j.o[go$externalize(method, Go$String)](selector.o, context.o);
			return (_struct$6 = j, new JQuery.Ptr(_struct$6.o, _struct$6.Jquery, _struct$6.Selector, _struct$6.Length, _struct$6.Context));
		} else if (_ref === 1) {
			_tuple$2 = (x$2 = ((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), (x$2 !== null && x$2.constructor === JQuery ? [x$2.go$val, true] : [new JQuery.Ptr(), false])); selector$1 = (_struct$7 = _tuple$2[0], new JQuery.Ptr(_struct$7.o, _struct$7.Jquery, _struct$7.Selector, _struct$7.Length, _struct$7.Context)); selOk$1 = _tuple$2[1];
			if (!selOk$1) {
				j.o = j.o[go$externalize(method, Go$String)](go$externalize(((0 < 0 || 0 >= i.length) ? go$throwRuntimeError("index out of range") : i.array[i.offset + 0]), go$emptyInterface));
				return (_struct$8 = j, new JQuery.Ptr(_struct$8.o, _struct$8.Jquery, _struct$8.Selector, _struct$8.Length, _struct$8.Context));
			}
			j.o = j.o[go$externalize(method, Go$String)](selector$1.o);
			return (_struct$9 = j, new JQuery.Ptr(_struct$9.o, _struct$9.Jquery, _struct$9.Selector, _struct$9.Length, _struct$9.Context));
		} else {
			console.log(" only 1 or 2 parameters allowed for method ", method);
			return (_struct$10 = j, new JQuery.Ptr(_struct$10.o, _struct$10.Jquery, _struct$10.Selector, _struct$10.Length, _struct$10.Context));
		}
	};
	JQuery.prototype.dom2args = function(method, i) { return this.go$val.dom2args(method, i); };
	JQuery.Ptr.prototype.dom1arg = function(method, i) {
		var _struct, j, _tuple, _struct$1, selector, selOk, _struct$2, _struct$3;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_tuple = (i !== null && i.constructor === JQuery ? [i.go$val, true] : [new JQuery.Ptr(), false]); selector = (_struct$1 = _tuple[0], new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context)); selOk = _tuple[1];
		if (!selOk) {
			j.o = j.o[go$externalize(method, Go$String)](go$externalize(i, go$emptyInterface));
			return (_struct$2 = j, new JQuery.Ptr(_struct$2.o, _struct$2.Jquery, _struct$2.Selector, _struct$2.Length, _struct$2.Context));
		}
		j.o = j.o[go$externalize(method, Go$String)](selector.o);
		return (_struct$3 = j, new JQuery.Ptr(_struct$3.o, _struct$3.Jquery, _struct$3.Selector, _struct$3.Length, _struct$3.Context));
	};
	JQuery.prototype.dom1arg = function(method, i) { return this.go$val.dom1arg(method, i); };
	JQuery.Ptr.prototype.Load = function(i) {
		var _struct, j, obj, _struct$1;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		j.o = (obj = j.o, obj.load.apply(obj, go$externalize(i, (go$sliceType(go$emptyInterface)))));
		return (_struct$1 = j, new JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context));
	};
	JQuery.prototype.Load = function(i) { return this.go$val.Load(i); };
	JQuery.Ptr.prototype.Serialize = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return go$internalize(j.o.serialize(), Go$String);
	};
	JQuery.prototype.Serialize = function() { return this.go$val.Serialize(); };
	JQuery.Ptr.prototype.SerializeArray = function() {
		var _struct, j;
		j = (_struct = this, new JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		return j.o.serializeArray();
	};
	JQuery.prototype.SerializeArray = function() { return this.go$val.SerializeArray(); };
	go$pkg.init = function() {
		JQuery.methods = [["Add", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["AddBack", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["AddClass", "", [go$emptyInterface], [JQuery], false, -1], ["After", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Append", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["AppendTo", "", [go$emptyInterface], [JQuery], false, -1], ["Attr", "", [Go$String], [Go$String], false, -1], ["Before", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Blur", "", [], [JQuery], false, -1], ["Children", "", [go$emptyInterface], [JQuery], false, -1], ["ClearQueue", "", [Go$String], [JQuery], false, -1], ["Clone", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Closest", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Contents", "", [], [JQuery], false, -1], ["Css", "", [Go$String], [Go$String], false, -1], ["CssArray", "", [(go$sliceType(Go$String))], [(go$mapType(Go$String, go$emptyInterface))], true, -1], ["Data", "", [Go$String], [go$emptyInterface], false, -1], ["Delay", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Dequeue", "", [Go$String], [JQuery], false, -1], ["Detach", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Each", "", [(go$funcType([Go$Int, go$emptyInterface], [go$emptyInterface], false))], [JQuery], false, -1], ["Empty", "", [], [JQuery], false, -1], ["End", "", [], [JQuery], false, -1], ["Eq", "", [Go$Int], [JQuery], false, -1], ["FadeIn", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["FadeOut", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Filter", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Find", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["First", "", [], [JQuery], false, -1], ["Focus", "", [], [JQuery], false, -1], ["Get", "", [(go$sliceType(go$emptyInterface))], [js.Object], true, -1], ["Has", "", [Go$String], [JQuery], false, -1], ["HasClass", "", [Go$String], [Go$Bool], false, -1], ["Height", "", [], [Go$Int], false, -1], ["Hide", "", [], [JQuery], false, -1], ["Html", "", [], [Go$String], false, -1], ["InnerHeight", "", [], [Go$Int], false, -1], ["InnerWidth", "", [], [Go$Int], false, -1], ["InsertAfter", "", [go$emptyInterface], [JQuery], false, -1], ["InsertBefore", "", [go$emptyInterface], [JQuery], false, -1], ["Is", "", [(go$sliceType(go$emptyInterface))], [Go$Bool], true, -1], ["Last", "", [], [JQuery], false, -1], ["Load", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Next", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["NextAll", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["NextUntil", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Not", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Off", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Offset", "", [], [JQueryCoordinates], false, -1], ["OffsetParent", "", [], [JQuery], false, -1], ["On", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["One", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["OuterHeight", "", [(go$sliceType(Go$Bool))], [Go$Int], true, -1], ["OuterWidth", "", [(go$sliceType(Go$Bool))], [Go$Int], true, -1], ["Parent", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Parents", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["ParentsUntil", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Position", "", [], [JQueryCoordinates], false, -1], ["Prepend", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["PrependTo", "", [go$emptyInterface], [JQuery], false, -1], ["Prev", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["PrevAll", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["PrevUntil", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Prop", "", [Go$String], [go$emptyInterface], false, -1], ["Ready", "", [(go$funcType([], [], false))], [JQuery], false, -1], ["Remove", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["RemoveAttr", "", [Go$String], [JQuery], false, -1], ["RemoveClass", "", [Go$String], [JQuery], false, -1], ["RemoveData", "", [Go$String], [JQuery], false, -1], ["RemoveProp", "", [Go$String], [JQuery], false, -1], ["ReplaceAll", "", [go$emptyInterface], [JQuery], false, -1], ["ReplaceWith", "", [go$emptyInterface], [JQuery], false, -1], ["Resize", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Scroll", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["ScrollLeft", "", [], [Go$Int], false, -1], ["ScrollTop", "", [], [Go$Int], false, -1], ["Select", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Serialize", "", [], [Go$String], false, -1], ["SerializeArray", "", [], [js.Object], false, -1], ["SetAttr", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["SetCss", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["SetData", "", [Go$String, go$emptyInterface], [JQuery], false, -1], ["SetHeight", "", [Go$String], [JQuery], false, -1], ["SetHtml", "", [go$emptyInterface], [JQuery], false, -1], ["SetOffset", "", [JQueryCoordinates], [JQuery], false, -1], ["SetProp", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["SetScrollLeft", "", [Go$Int], [JQuery], false, -1], ["SetScrollTop", "", [Go$Int], [JQuery], false, -1], ["SetText", "", [go$emptyInterface], [JQuery], false, -1], ["SetVal", "", [go$emptyInterface], [JQuery], false, -1], ["SetWidth", "", [go$emptyInterface], [JQuery], false, -1], ["Show", "", [], [JQuery], false, -1], ["Siblings", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Slice", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Stop", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Submit", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Text", "", [], [Go$String], false, -1], ["ToArray", "", [], [(go$sliceType(go$emptyInterface))], false, -1], ["Toggle", "", [Go$Bool], [JQuery], false, -1], ["ToggleClass", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Trigger", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Underlying", "", [], [js.Object], false, -1], ["Unwrap", "", [], [JQuery], false, -1], ["Val", "", [], [Go$String], false, -1], ["Width", "", [], [Go$Int], false, -1], ["Wrap", "", [go$emptyInterface], [JQuery], false, -1], ["WrapAll", "", [go$emptyInterface], [JQuery], false, -1], ["WrapInner", "", [go$emptyInterface], [JQuery], false, -1], ["dom1arg", "github.com/gopherjs/jquery", [Go$String, go$emptyInterface], [JQuery], false, -1], ["dom2args", "github.com/gopherjs/jquery", [Go$String, (go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["events", "github.com/gopherjs/jquery", [Go$String, (go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["handleEvent", "github.com/gopherjs/jquery", [Go$String, (go$sliceType(go$emptyInterface))], [JQuery], true, -1]];
		(go$ptrType(JQuery)).methods = [["Add", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["AddBack", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["AddClass", "", [go$emptyInterface], [JQuery], false, -1], ["After", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Append", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["AppendTo", "", [go$emptyInterface], [JQuery], false, -1], ["Attr", "", [Go$String], [Go$String], false, -1], ["Before", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Blur", "", [], [JQuery], false, -1], ["Children", "", [go$emptyInterface], [JQuery], false, -1], ["ClearQueue", "", [Go$String], [JQuery], false, -1], ["Clone", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Closest", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Contents", "", [], [JQuery], false, -1], ["Css", "", [Go$String], [Go$String], false, -1], ["CssArray", "", [(go$sliceType(Go$String))], [(go$mapType(Go$String, go$emptyInterface))], true, -1], ["Data", "", [Go$String], [go$emptyInterface], false, -1], ["Delay", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Dequeue", "", [Go$String], [JQuery], false, -1], ["Detach", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Each", "", [(go$funcType([Go$Int, go$emptyInterface], [go$emptyInterface], false))], [JQuery], false, -1], ["Empty", "", [], [JQuery], false, -1], ["End", "", [], [JQuery], false, -1], ["Eq", "", [Go$Int], [JQuery], false, -1], ["FadeIn", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["FadeOut", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Filter", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Find", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["First", "", [], [JQuery], false, -1], ["Focus", "", [], [JQuery], false, -1], ["Get", "", [(go$sliceType(go$emptyInterface))], [js.Object], true, -1], ["Has", "", [Go$String], [JQuery], false, -1], ["HasClass", "", [Go$String], [Go$Bool], false, -1], ["Height", "", [], [Go$Int], false, -1], ["Hide", "", [], [JQuery], false, -1], ["Html", "", [], [Go$String], false, -1], ["InnerHeight", "", [], [Go$Int], false, -1], ["InnerWidth", "", [], [Go$Int], false, -1], ["InsertAfter", "", [go$emptyInterface], [JQuery], false, -1], ["InsertBefore", "", [go$emptyInterface], [JQuery], false, -1], ["Is", "", [(go$sliceType(go$emptyInterface))], [Go$Bool], true, -1], ["Last", "", [], [JQuery], false, -1], ["Load", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Next", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["NextAll", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["NextUntil", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Not", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Off", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Offset", "", [], [JQueryCoordinates], false, -1], ["OffsetParent", "", [], [JQuery], false, -1], ["On", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["One", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["OuterHeight", "", [(go$sliceType(Go$Bool))], [Go$Int], true, -1], ["OuterWidth", "", [(go$sliceType(Go$Bool))], [Go$Int], true, -1], ["Parent", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Parents", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["ParentsUntil", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Position", "", [], [JQueryCoordinates], false, -1], ["Prepend", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["PrependTo", "", [go$emptyInterface], [JQuery], false, -1], ["Prev", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["PrevAll", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["PrevUntil", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Prop", "", [Go$String], [go$emptyInterface], false, -1], ["Ready", "", [(go$funcType([], [], false))], [JQuery], false, -1], ["Remove", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["RemoveAttr", "", [Go$String], [JQuery], false, -1], ["RemoveClass", "", [Go$String], [JQuery], false, -1], ["RemoveData", "", [Go$String], [JQuery], false, -1], ["RemoveProp", "", [Go$String], [JQuery], false, -1], ["ReplaceAll", "", [go$emptyInterface], [JQuery], false, -1], ["ReplaceWith", "", [go$emptyInterface], [JQuery], false, -1], ["Resize", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Scroll", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["ScrollLeft", "", [], [Go$Int], false, -1], ["ScrollTop", "", [], [Go$Int], false, -1], ["Select", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Serialize", "", [], [Go$String], false, -1], ["SerializeArray", "", [], [js.Object], false, -1], ["SetAttr", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["SetCss", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["SetData", "", [Go$String, go$emptyInterface], [JQuery], false, -1], ["SetHeight", "", [Go$String], [JQuery], false, -1], ["SetHtml", "", [go$emptyInterface], [JQuery], false, -1], ["SetOffset", "", [JQueryCoordinates], [JQuery], false, -1], ["SetProp", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["SetScrollLeft", "", [Go$Int], [JQuery], false, -1], ["SetScrollTop", "", [Go$Int], [JQuery], false, -1], ["SetText", "", [go$emptyInterface], [JQuery], false, -1], ["SetVal", "", [go$emptyInterface], [JQuery], false, -1], ["SetWidth", "", [go$emptyInterface], [JQuery], false, -1], ["Show", "", [], [JQuery], false, -1], ["Siblings", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Slice", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Stop", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Submit", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Text", "", [], [Go$String], false, -1], ["ToArray", "", [], [(go$sliceType(go$emptyInterface))], false, -1], ["Toggle", "", [Go$Bool], [JQuery], false, -1], ["ToggleClass", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Trigger", "", [(go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["Underlying", "", [], [js.Object], false, -1], ["Unwrap", "", [], [JQuery], false, -1], ["Val", "", [], [Go$String], false, -1], ["Width", "", [], [Go$Int], false, -1], ["Wrap", "", [go$emptyInterface], [JQuery], false, -1], ["WrapAll", "", [go$emptyInterface], [JQuery], false, -1], ["WrapInner", "", [go$emptyInterface], [JQuery], false, -1], ["dom1arg", "github.com/gopherjs/jquery", [Go$String, go$emptyInterface], [JQuery], false, -1], ["dom2args", "github.com/gopherjs/jquery", [Go$String, (go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["events", "github.com/gopherjs/jquery", [Go$String, (go$sliceType(go$emptyInterface))], [JQuery], true, -1], ["handleEvent", "github.com/gopherjs/jquery", [Go$String, (go$sliceType(go$emptyInterface))], [JQuery], true, -1]];
		JQuery.init([["o", "o", "github.com/gopherjs/jquery", js.Object, ""], ["Jquery", "Jquery", "", Go$String, "js:\"jquery\""], ["Selector", "Selector", "", Go$String, "js:\"selector\""], ["Length", "Length", "", Go$String, "js:\"length\""], ["Context", "Context", "", Go$String, "js:\"context\""]]);
		Event.methods = [["Bool", "", [], [Go$Bool], false, 0], ["Call", "", [Go$String, (go$sliceType(go$emptyInterface))], [js.Object], true, 0], ["Delete", "", [Go$String], [], false, 0], ["Float", "", [], [Go$Float64], false, 0], ["Get", "", [Go$String], [js.Object], false, 0], ["Index", "", [Go$Int], [js.Object], false, 0], ["Int", "", [], [Go$Int], false, 0], ["Int64", "", [], [Go$Int64], false, 0], ["Interface", "", [], [go$emptyInterface], false, 0], ["Invoke", "", [(go$sliceType(go$emptyInterface))], [js.Object], true, 0], ["IsNull", "", [], [Go$Bool], false, 0], ["IsUndefined", "", [], [Go$Bool], false, 0], ["Length", "", [], [Go$Int], false, 0], ["New", "", [(go$sliceType(go$emptyInterface))], [js.Object], true, 0], ["Set", "", [Go$String, go$emptyInterface], [], false, 0], ["SetIndex", "", [Go$Int, go$emptyInterface], [], false, 0], ["Str", "", [], [Go$String], false, 0], ["Uint64", "", [], [Go$Uint64], false, 0], ["Unsafe", "", [], [Go$Uintptr], false, 0]];
		(go$ptrType(Event)).methods = [["Bool", "", [], [Go$Bool], false, 0], ["Call", "", [Go$String, (go$sliceType(go$emptyInterface))], [js.Object], true, 0], ["Delete", "", [Go$String], [], false, 0], ["Float", "", [], [Go$Float64], false, 0], ["Get", "", [Go$String], [js.Object], false, 0], ["Index", "", [Go$Int], [js.Object], false, 0], ["Int", "", [], [Go$Int], false, 0], ["Int64", "", [], [Go$Int64], false, 0], ["Interface", "", [], [go$emptyInterface], false, 0], ["Invoke", "", [(go$sliceType(go$emptyInterface))], [js.Object], true, 0], ["IsDefaultPrevented", "", [], [Go$Bool], false, -1], ["IsImmediatePropogationStopped", "", [], [Go$Bool], false, -1], ["IsNull", "", [], [Go$Bool], false, 0], ["IsPropagationStopped", "", [], [Go$Bool], false, -1], ["IsUndefined", "", [], [Go$Bool], false, 0], ["Length", "", [], [Go$Int], false, 0], ["New", "", [(go$sliceType(go$emptyInterface))], [js.Object], true, 0], ["PreventDefault", "", [], [], false, -1], ["Set", "", [Go$String, go$emptyInterface], [], false, 0], ["SetIndex", "", [Go$Int, go$emptyInterface], [], false, 0], ["StopImmediatePropagation", "", [], [], false, -1], ["StopPropagation", "", [], [], false, -1], ["Str", "", [], [Go$String], false, 0], ["Uint64", "", [], [Go$Uint64], false, 0], ["Unsafe", "", [], [Go$Uintptr], false, 0]];
		Event.init([["Object", "", "", js.Object, ""], ["KeyCode", "KeyCode", "", Go$Int, "js:\"keyCode\""], ["Target", "Target", "", js.Object, "js:\"target\""], ["CurrentTarget", "CurrentTarget", "", js.Object, "js:\"currentTarget\""], ["DelegateTarget", "DelegateTarget", "", js.Object, "js:\"delegateTarget\""], ["RelatedTarget", "RelatedTarget", "", js.Object, "js:\"relatedTarget\""], ["Data", "Data", "", js.Object, "js:\"data\""], ["Result", "Result", "", js.Object, "js:\"result\""], ["Which", "Which", "", Go$Int, "js:\"which\""], ["Namespace", "Namespace", "", Go$String, "js:\"namespace\""], ["MetaKey", "MetaKey", "", Go$Bool, "js:\"metaKey\""], ["PageX", "PageX", "", Go$Int, "js:\"pageX\""], ["PageY", "PageY", "", Go$Int, "js:\"pageY\""], ["Type", "Type", "", Go$String, "js:\"type\""]]);
		JQueryCoordinates.init([["Left", "Left", "", Go$Int, ""], ["Top", "Top", "", Go$Int, ""]]);
	}
	return go$pkg;
})();
go$packages["errors"] = (function() {
	var go$pkg = {}, errorString, New;
	errorString = go$pkg.errorString = go$newType(0, "Struct", "errors.errorString", "errorString", "errors", function(s_) {
		this.go$val = this;
		this.s = s_ !== undefined ? s_ : "";
	});
	New = go$pkg.New = function(text) {
		return new errorString.Ptr(text);
	};
	errorString.Ptr.prototype.Error = function() {
		var e;
		e = this;
		return e.s;
	};
	errorString.prototype.Error = function() { return this.go$val.Error(); };
	go$pkg.init = function() {
		(go$ptrType(errorString)).methods = [["Error", "", [], [Go$String], false, -1]];
		errorString.init([["s", "s", "errors", Go$String, ""]]);
	}
	return go$pkg;
})();
go$packages["sync/atomic"] = (function() {
	var go$pkg = {}, CompareAndSwapInt32, AddInt32, LoadUint32, StoreInt32, StoreUint32;
	CompareAndSwapInt32 = go$pkg.CompareAndSwapInt32 = function(addr, old, new$1) {
		if (addr.go$get() === old) {
			addr.go$set(new$1);
			return true;
		}
		return false;
	};
	AddInt32 = go$pkg.AddInt32 = function(addr, delta) {
		var new$1;
		new$1 = addr.go$get() + delta >> 0;
		addr.go$set(new$1);
		return new$1;
	};
	LoadUint32 = go$pkg.LoadUint32 = function(addr) {
		return addr.go$get();
	};
	StoreInt32 = go$pkg.StoreInt32 = function(addr, val) {
		addr.go$set(val);
	};
	StoreUint32 = go$pkg.StoreUint32 = function(addr, val) {
		addr.go$set(val);
	};
	go$pkg.init = function() {
	}
	return go$pkg;
})();
go$packages["sync"] = (function() {
	var go$pkg = {}, atomic = go$packages["sync/atomic"], Mutex, Locker, Once, RWMutex, rlocker, runtime_Syncsemcheck, runtime_Semacquire, runtime_Semrelease;
	Mutex = go$pkg.Mutex = go$newType(0, "Struct", "sync.Mutex", "Mutex", "sync", function(state_, sema_) {
		this.go$val = this;
		this.state = state_ !== undefined ? state_ : 0;
		this.sema = sema_ !== undefined ? sema_ : 0;
	});
	Locker = go$pkg.Locker = go$newType(8, "Interface", "sync.Locker", "Locker", "sync", null);
	Once = go$pkg.Once = go$newType(0, "Struct", "sync.Once", "Once", "sync", function(m_, done_) {
		this.go$val = this;
		this.m = m_ !== undefined ? m_ : new Mutex.Ptr();
		this.done = done_ !== undefined ? done_ : 0;
	});
	RWMutex = go$pkg.RWMutex = go$newType(0, "Struct", "sync.RWMutex", "RWMutex", "sync", function(w_, writerSem_, readerSem_, readerCount_, readerWait_) {
		this.go$val = this;
		this.w = w_ !== undefined ? w_ : new Mutex.Ptr();
		this.writerSem = writerSem_ !== undefined ? writerSem_ : 0;
		this.readerSem = readerSem_ !== undefined ? readerSem_ : 0;
		this.readerCount = readerCount_ !== undefined ? readerCount_ : 0;
		this.readerWait = readerWait_ !== undefined ? readerWait_ : 0;
	});
	rlocker = go$pkg.rlocker = go$newType(0, "Struct", "sync.rlocker", "rlocker", "sync", function(w_, writerSem_, readerSem_, readerCount_, readerWait_) {
		this.go$val = this;
		this.w = w_ !== undefined ? w_ : new Mutex.Ptr();
		this.writerSem = writerSem_ !== undefined ? writerSem_ : 0;
		this.readerSem = readerSem_ !== undefined ? readerSem_ : 0;
		this.readerCount = readerCount_ !== undefined ? readerCount_ : 0;
		this.readerWait = readerWait_ !== undefined ? readerWait_ : 0;
	});
	runtime_Syncsemcheck = function(size) {
	};
	Mutex.Ptr.prototype.Lock = function() {
		var m, v, awoke, old, new$1, v$1, v$2;
		m = this;
		if (atomic.CompareAndSwapInt32(new (go$ptrType(Go$Int32))(function() { return m.state; }, function(v) { m.state = v;; }), 0, 1)) {
			return;
		}
		awoke = false;
		while (true) {
			old = m.state;
			new$1 = old | 1;
			if (!(((old & 1) === 0))) {
				new$1 = old + 4 >> 0;
			}
			if (awoke) {
				new$1 = new$1 & ~2;
			}
			if (atomic.CompareAndSwapInt32(new (go$ptrType(Go$Int32))(function() { return m.state; }, function(v$1) { m.state = v$1;; }), old, new$1)) {
				if ((old & 1) === 0) {
					break;
				}
				runtime_Semacquire(new (go$ptrType(Go$Uint32))(function() { return m.sema; }, function(v$2) { m.sema = v$2;; }));
				awoke = true;
			}
		}
	};
	Mutex.prototype.Lock = function() { return this.go$val.Lock(); };
	Mutex.Ptr.prototype.Unlock = function() {
		var m, v, new$1, old, v$1, v$2;
		m = this;
		new$1 = atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return m.state; }, function(v) { m.state = v;; }), -1);
		if ((((new$1 + 1 >> 0)) & 1) === 0) {
			throw go$panic(new Go$String("sync: unlock of unlocked mutex"));
		}
		old = new$1;
		while (true) {
			if (((old >> 2 >> 0) === 0) || !(((old & 3) === 0))) {
				return;
			}
			new$1 = ((old - 4 >> 0)) | 2;
			if (atomic.CompareAndSwapInt32(new (go$ptrType(Go$Int32))(function() { return m.state; }, function(v$1) { m.state = v$1;; }), old, new$1)) {
				runtime_Semrelease(new (go$ptrType(Go$Uint32))(function() { return m.sema; }, function(v$2) { m.sema = v$2;; }));
				return;
			}
			old = m.state;
		}
	};
	Mutex.prototype.Unlock = function() { return this.go$val.Unlock(); };
	Once.Ptr.prototype.Do = function(f) {
		var o, v, v$1;
		var go$deferred = [];
		try {
			o = this;
			if (atomic.LoadUint32(new (go$ptrType(Go$Uint32))(function() { return o.done; }, function(v) { o.done = v;; })) === 1) {
				return;
			}
			o.m.Lock();
			go$deferred.push({ recv: o.m, method: "Unlock", args: [] });
			if (o.done === 0) {
				f();
				atomic.StoreUint32(new (go$ptrType(Go$Uint32))(function() { return o.done; }, function(v$1) { o.done = v$1;; }), 1);
			}
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
		}
	};
	Once.prototype.Do = function(f) { return this.go$val.Do(f); };
	runtime_Semacquire = function() {
		throw go$panic("Native function not implemented: runtime_Semacquire");
	};
	runtime_Semrelease = function() {
		throw go$panic("Native function not implemented: runtime_Semrelease");
	};
	RWMutex.Ptr.prototype.RLock = function() {
		var rw, v, v$1;
		rw = this;
		if (atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return rw.readerCount; }, function(v) { rw.readerCount = v;; }), 1) < 0) {
			runtime_Semacquire(new (go$ptrType(Go$Uint32))(function() { return rw.readerSem; }, function(v$1) { rw.readerSem = v$1;; }));
		}
	};
	RWMutex.prototype.RLock = function() { return this.go$val.RLock(); };
	RWMutex.Ptr.prototype.RUnlock = function() {
		var rw, v, v$1, v$2;
		rw = this;
		if (atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return rw.readerCount; }, function(v) { rw.readerCount = v;; }), -1) < 0) {
			if (atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return rw.readerWait; }, function(v$1) { rw.readerWait = v$1;; }), -1) === 0) {
				runtime_Semrelease(new (go$ptrType(Go$Uint32))(function() { return rw.writerSem; }, function(v$2) { rw.writerSem = v$2;; }));
			}
		}
	};
	RWMutex.prototype.RUnlock = function() { return this.go$val.RUnlock(); };
	RWMutex.Ptr.prototype.Lock = function() {
		var rw, v, r, v$1, v$2;
		rw = this;
		rw.w.Lock();
		r = atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return rw.readerCount; }, function(v) { rw.readerCount = v;; }), -1073741824) + 1073741824 >> 0;
		if (!((r === 0)) && !((atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return rw.readerWait; }, function(v$1) { rw.readerWait = v$1;; }), r) === 0))) {
			runtime_Semacquire(new (go$ptrType(Go$Uint32))(function() { return rw.writerSem; }, function(v$2) { rw.writerSem = v$2;; }));
		}
	};
	RWMutex.prototype.Lock = function() { return this.go$val.Lock(); };
	RWMutex.Ptr.prototype.Unlock = function() {
		var rw, v, r, i, v$1;
		rw = this;
		r = atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return rw.readerCount; }, function(v) { rw.readerCount = v;; }), 1073741824);
		i = 0;
		while (i < (r >> 0)) {
			runtime_Semrelease(new (go$ptrType(Go$Uint32))(function() { return rw.readerSem; }, function(v$1) { rw.readerSem = v$1;; }));
			i = i + 1 >> 0;
		}
		rw.w.Unlock();
	};
	RWMutex.prototype.Unlock = function() { return this.go$val.Unlock(); };
	RWMutex.Ptr.prototype.RLocker = function() {
		var rw, _struct, _struct$1;
		rw = this;
		return (_struct = rw, new rlocker.Ptr((_struct$1 = _struct.w, new Mutex.Ptr(_struct$1.state, _struct$1.sema)), _struct.writerSem, _struct.readerSem, _struct.readerCount, _struct.readerWait));
	};
	RWMutex.prototype.RLocker = function() { return this.go$val.RLocker(); };
	rlocker.Ptr.prototype.Lock = function() {
		var r, _struct, _struct$1;
		r = this;
		(_struct = r, new RWMutex.Ptr((_struct$1 = _struct.w, new Mutex.Ptr(_struct$1.state, _struct$1.sema)), _struct.writerSem, _struct.readerSem, _struct.readerCount, _struct.readerWait)).RLock();
	};
	rlocker.prototype.Lock = function() { return this.go$val.Lock(); };
	rlocker.Ptr.prototype.Unlock = function() {
		var r, _struct, _struct$1;
		r = this;
		(_struct = r, new RWMutex.Ptr((_struct$1 = _struct.w, new Mutex.Ptr(_struct$1.state, _struct$1.sema)), _struct.writerSem, _struct.readerSem, _struct.readerCount, _struct.readerWait)).RUnlock();
	};
	rlocker.prototype.Unlock = function() { return this.go$val.Unlock(); };
	go$pkg.init = function() {
		(go$ptrType(Mutex)).methods = [["Lock", "", [], [], false, -1], ["Unlock", "", [], [], false, -1]];
		Mutex.init([["state", "state", "sync", Go$Int32, ""], ["sema", "sema", "sync", Go$Uint32, ""]]);
		Locker.init([["Lock", "", (go$funcType([], [], false))], ["Unlock", "", (go$funcType([], [], false))]]);
		(go$ptrType(Once)).methods = [["Do", "", [(go$funcType([], [], false))], [], false, -1]];
		Once.init([["m", "m", "sync", Mutex, ""], ["done", "done", "sync", Go$Uint32, ""]]);
		(go$ptrType(RWMutex)).methods = [["Lock", "", [], [], false, -1], ["RLock", "", [], [], false, -1], ["RLocker", "", [], [Locker], false, -1], ["RUnlock", "", [], [], false, -1], ["Unlock", "", [], [], false, -1]];
		RWMutex.init([["w", "w", "sync", Mutex, ""], ["writerSem", "writerSem", "sync", Go$Uint32, ""], ["readerSem", "readerSem", "sync", Go$Uint32, ""], ["readerCount", "readerCount", "sync", Go$Int32, ""], ["readerWait", "readerWait", "sync", Go$Int32, ""]]);
		(go$ptrType(rlocker)).methods = [["Lock", "", [], [], false, -1], ["Unlock", "", [], [], false, -1]];
		rlocker.init([["w", "w", "sync", Mutex, ""], ["writerSem", "writerSem", "sync", Go$Uint32, ""], ["readerSem", "readerSem", "sync", Go$Uint32, ""], ["readerCount", "readerCount", "sync", Go$Int32, ""], ["readerWait", "readerWait", "sync", Go$Int32, ""]]);
		var s;
		s = go$makeNativeArray("Uintptr", 3, function() { return 0; });
		runtime_Syncsemcheck(12);
	}
	return go$pkg;
})();
go$packages["io"] = (function() {
	var go$pkg = {}, errors = go$packages["errors"], sync = go$packages["sync"], RuneReader, errWhence, errOffset;
	RuneReader = go$pkg.RuneReader = go$newType(8, "Interface", "io.RuneReader", "RuneReader", "io", null);
	go$pkg.init = function() {
		RuneReader.init([["ReadRune", "", (go$funcType([], [Go$Int32, Go$Int, go$error], false))]]);
		go$pkg.ErrShortWrite = errors.New("short write");
		go$pkg.ErrShortBuffer = errors.New("short buffer");
		go$pkg.EOF = errors.New("EOF");
		go$pkg.ErrUnexpectedEOF = errors.New("unexpected EOF");
		go$pkg.ErrNoProgress = errors.New("multiple Read calls return no data or error");
		errWhence = errors.New("Seek: invalid whence");
		errOffset = errors.New("Seek: invalid offset");
		go$pkg.ErrClosedPipe = errors.New("io: read/write on closed pipe");
	}
	return go$pkg;
})();
go$packages["unicode"] = (function() {
	var go$pkg = {}, CaseRange, d, to, To, ToLower, _CaseRanges;
	CaseRange = go$pkg.CaseRange = go$newType(0, "Struct", "unicode.CaseRange", "CaseRange", "unicode", function(Lo_, Hi_, Delta_) {
		this.go$val = this;
		this.Lo = Lo_ !== undefined ? Lo_ : 0;
		this.Hi = Hi_ !== undefined ? Hi_ : 0;
		this.Delta = Delta_ !== undefined ? Delta_ : go$makeNativeArray("Int32", 3, function() { return 0; });
	});
	d = go$pkg.d = go$newType(12, "Array", "unicode.d", "d", "unicode", null);
	to = function(_case, r, caseRange) {
		var lo, hi, _q, m, _struct, cr, delta;
		if (_case < 0 || 3 <= _case) {
			return 65533;
		}
		lo = 0;
		hi = caseRange.length;
		while (lo < hi) {
			m = lo + (_q = ((hi - lo >> 0)) / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0;
			cr = (_struct = ((m < 0 || m >= caseRange.length) ? go$throwRuntimeError("index out of range") : caseRange.array[caseRange.offset + m]), new CaseRange.Ptr(_struct.Lo, _struct.Hi, go$mapArray(_struct.Delta, function(entry) { return entry; })));
			if ((cr.Lo >> 0) <= r && r <= (cr.Hi >> 0)) {
				delta = cr.Delta[_case];
				if (delta > 1114111) {
					return (cr.Lo >> 0) + (((((r - (cr.Lo >> 0) >> 0)) & ~1) | ((_case & 1) >> 0))) >> 0;
				}
				return r + delta >> 0;
			}
			if (r < (cr.Lo >> 0)) {
				hi = m;
			} else {
				lo = m + 1 >> 0;
			}
		}
		return r;
	};
	To = go$pkg.To = function(_case, r) {
		return to(_case, r, go$pkg.CaseRanges);
	};
	ToLower = go$pkg.ToLower = function(r) {
		if (r <= 127) {
			if (65 <= r && r <= 90) {
				r = r + 32 >> 0;
			}
			return r;
		}
		return To(1, r);
	};
	go$pkg.init = function() {
		CaseRange.init([["Lo", "Lo", "", Go$Uint32, ""], ["Hi", "Hi", "", Go$Uint32, ""], ["Delta", "Delta", "", d, ""]]);
		d.init(Go$Int32, 3);
		_CaseRanges = new (go$sliceType(CaseRange))([new CaseRange.Ptr(65, 90, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(97, 122, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(181, 181, go$toNativeArray("Int32", [743, 0, 743])), new CaseRange.Ptr(192, 214, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(216, 222, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(224, 246, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(248, 254, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(255, 255, go$toNativeArray("Int32", [121, 0, 121])), new CaseRange.Ptr(256, 303, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(304, 304, go$toNativeArray("Int32", [0, -199, 0])), new CaseRange.Ptr(305, 305, go$toNativeArray("Int32", [-232, 0, -232])), new CaseRange.Ptr(306, 311, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(313, 328, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(330, 375, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(376, 376, go$toNativeArray("Int32", [0, -121, 0])), new CaseRange.Ptr(377, 382, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(383, 383, go$toNativeArray("Int32", [-300, 0, -300])), new CaseRange.Ptr(384, 384, go$toNativeArray("Int32", [195, 0, 195])), new CaseRange.Ptr(385, 385, go$toNativeArray("Int32", [0, 210, 0])), new CaseRange.Ptr(386, 389, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(390, 390, go$toNativeArray("Int32", [0, 206, 0])), new CaseRange.Ptr(391, 392, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(393, 394, go$toNativeArray("Int32", [0, 205, 0])), new CaseRange.Ptr(395, 396, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(398, 398, go$toNativeArray("Int32", [0, 79, 0])), new CaseRange.Ptr(399, 399, go$toNativeArray("Int32", [0, 202, 0])), new CaseRange.Ptr(400, 400, go$toNativeArray("Int32", [0, 203, 0])), new CaseRange.Ptr(401, 402, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(403, 403, go$toNativeArray("Int32", [0, 205, 0])), new CaseRange.Ptr(404, 404, go$toNativeArray("Int32", [0, 207, 0])), new CaseRange.Ptr(405, 405, go$toNativeArray("Int32", [97, 0, 97])), new CaseRange.Ptr(406, 406, go$toNativeArray("Int32", [0, 211, 0])), new CaseRange.Ptr(407, 407, go$toNativeArray("Int32", [0, 209, 0])), new CaseRange.Ptr(408, 409, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(410, 410, go$toNativeArray("Int32", [163, 0, 163])), new CaseRange.Ptr(412, 412, go$toNativeArray("Int32", [0, 211, 0])), new CaseRange.Ptr(413, 413, go$toNativeArray("Int32", [0, 213, 0])), new CaseRange.Ptr(414, 414, go$toNativeArray("Int32", [130, 0, 130])), new CaseRange.Ptr(415, 415, go$toNativeArray("Int32", [0, 214, 0])), new CaseRange.Ptr(416, 421, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(422, 422, go$toNativeArray("Int32", [0, 218, 0])), new CaseRange.Ptr(423, 424, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(425, 425, go$toNativeArray("Int32", [0, 218, 0])), new CaseRange.Ptr(428, 429, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(430, 430, go$toNativeArray("Int32", [0, 218, 0])), new CaseRange.Ptr(431, 432, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(433, 434, go$toNativeArray("Int32", [0, 217, 0])), new CaseRange.Ptr(435, 438, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(439, 439, go$toNativeArray("Int32", [0, 219, 0])), new CaseRange.Ptr(440, 441, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(444, 445, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(447, 447, go$toNativeArray("Int32", [56, 0, 56])), new CaseRange.Ptr(452, 452, go$toNativeArray("Int32", [0, 2, 1])), new CaseRange.Ptr(453, 453, go$toNativeArray("Int32", [-1, 1, 0])), new CaseRange.Ptr(454, 454, go$toNativeArray("Int32", [-2, 0, -1])), new CaseRange.Ptr(455, 455, go$toNativeArray("Int32", [0, 2, 1])), new CaseRange.Ptr(456, 456, go$toNativeArray("Int32", [-1, 1, 0])), new CaseRange.Ptr(457, 457, go$toNativeArray("Int32", [-2, 0, -1])), new CaseRange.Ptr(458, 458, go$toNativeArray("Int32", [0, 2, 1])), new CaseRange.Ptr(459, 459, go$toNativeArray("Int32", [-1, 1, 0])), new CaseRange.Ptr(460, 460, go$toNativeArray("Int32", [-2, 0, -1])), new CaseRange.Ptr(461, 476, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(477, 477, go$toNativeArray("Int32", [-79, 0, -79])), new CaseRange.Ptr(478, 495, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(497, 497, go$toNativeArray("Int32", [0, 2, 1])), new CaseRange.Ptr(498, 498, go$toNativeArray("Int32", [-1, 1, 0])), new CaseRange.Ptr(499, 499, go$toNativeArray("Int32", [-2, 0, -1])), new CaseRange.Ptr(500, 501, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(502, 502, go$toNativeArray("Int32", [0, -97, 0])), new CaseRange.Ptr(503, 503, go$toNativeArray("Int32", [0, -56, 0])), new CaseRange.Ptr(504, 543, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(544, 544, go$toNativeArray("Int32", [0, -130, 0])), new CaseRange.Ptr(546, 563, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(570, 570, go$toNativeArray("Int32", [0, 10795, 0])), new CaseRange.Ptr(571, 572, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(573, 573, go$toNativeArray("Int32", [0, -163, 0])), new CaseRange.Ptr(574, 574, go$toNativeArray("Int32", [0, 10792, 0])), new CaseRange.Ptr(575, 576, go$toNativeArray("Int32", [10815, 0, 10815])), new CaseRange.Ptr(577, 578, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(579, 579, go$toNativeArray("Int32", [0, -195, 0])), new CaseRange.Ptr(580, 580, go$toNativeArray("Int32", [0, 69, 0])), new CaseRange.Ptr(581, 581, go$toNativeArray("Int32", [0, 71, 0])), new CaseRange.Ptr(582, 591, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(592, 592, go$toNativeArray("Int32", [10783, 0, 10783])), new CaseRange.Ptr(593, 593, go$toNativeArray("Int32", [10780, 0, 10780])), new CaseRange.Ptr(594, 594, go$toNativeArray("Int32", [10782, 0, 10782])), new CaseRange.Ptr(595, 595, go$toNativeArray("Int32", [-210, 0, -210])), new CaseRange.Ptr(596, 596, go$toNativeArray("Int32", [-206, 0, -206])), new CaseRange.Ptr(598, 599, go$toNativeArray("Int32", [-205, 0, -205])), new CaseRange.Ptr(601, 601, go$toNativeArray("Int32", [-202, 0, -202])), new CaseRange.Ptr(603, 603, go$toNativeArray("Int32", [-203, 0, -203])), new CaseRange.Ptr(608, 608, go$toNativeArray("Int32", [-205, 0, -205])), new CaseRange.Ptr(611, 611, go$toNativeArray("Int32", [-207, 0, -207])), new CaseRange.Ptr(613, 613, go$toNativeArray("Int32", [42280, 0, 42280])), new CaseRange.Ptr(614, 614, go$toNativeArray("Int32", [42308, 0, 42308])), new CaseRange.Ptr(616, 616, go$toNativeArray("Int32", [-209, 0, -209])), new CaseRange.Ptr(617, 617, go$toNativeArray("Int32", [-211, 0, -211])), new CaseRange.Ptr(619, 619, go$toNativeArray("Int32", [10743, 0, 10743])), new CaseRange.Ptr(623, 623, go$toNativeArray("Int32", [-211, 0, -211])), new CaseRange.Ptr(625, 625, go$toNativeArray("Int32", [10749, 0, 10749])), new CaseRange.Ptr(626, 626, go$toNativeArray("Int32", [-213, 0, -213])), new CaseRange.Ptr(629, 629, go$toNativeArray("Int32", [-214, 0, -214])), new CaseRange.Ptr(637, 637, go$toNativeArray("Int32", [10727, 0, 10727])), new CaseRange.Ptr(640, 640, go$toNativeArray("Int32", [-218, 0, -218])), new CaseRange.Ptr(643, 643, go$toNativeArray("Int32", [-218, 0, -218])), new CaseRange.Ptr(648, 648, go$toNativeArray("Int32", [-218, 0, -218])), new CaseRange.Ptr(649, 649, go$toNativeArray("Int32", [-69, 0, -69])), new CaseRange.Ptr(650, 651, go$toNativeArray("Int32", [-217, 0, -217])), new CaseRange.Ptr(652, 652, go$toNativeArray("Int32", [-71, 0, -71])), new CaseRange.Ptr(658, 658, go$toNativeArray("Int32", [-219, 0, -219])), new CaseRange.Ptr(837, 837, go$toNativeArray("Int32", [84, 0, 84])), new CaseRange.Ptr(880, 883, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(886, 887, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(891, 893, go$toNativeArray("Int32", [130, 0, 130])), new CaseRange.Ptr(902, 902, go$toNativeArray("Int32", [0, 38, 0])), new CaseRange.Ptr(904, 906, go$toNativeArray("Int32", [0, 37, 0])), new CaseRange.Ptr(908, 908, go$toNativeArray("Int32", [0, 64, 0])), new CaseRange.Ptr(910, 911, go$toNativeArray("Int32", [0, 63, 0])), new CaseRange.Ptr(913, 929, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(931, 939, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(940, 940, go$toNativeArray("Int32", [-38, 0, -38])), new CaseRange.Ptr(941, 943, go$toNativeArray("Int32", [-37, 0, -37])), new CaseRange.Ptr(945, 961, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(962, 962, go$toNativeArray("Int32", [-31, 0, -31])), new CaseRange.Ptr(963, 971, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(972, 972, go$toNativeArray("Int32", [-64, 0, -64])), new CaseRange.Ptr(973, 974, go$toNativeArray("Int32", [-63, 0, -63])), new CaseRange.Ptr(975, 975, go$toNativeArray("Int32", [0, 8, 0])), new CaseRange.Ptr(976, 976, go$toNativeArray("Int32", [-62, 0, -62])), new CaseRange.Ptr(977, 977, go$toNativeArray("Int32", [-57, 0, -57])), new CaseRange.Ptr(981, 981, go$toNativeArray("Int32", [-47, 0, -47])), new CaseRange.Ptr(982, 982, go$toNativeArray("Int32", [-54, 0, -54])), new CaseRange.Ptr(983, 983, go$toNativeArray("Int32", [-8, 0, -8])), new CaseRange.Ptr(984, 1007, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1008, 1008, go$toNativeArray("Int32", [-86, 0, -86])), new CaseRange.Ptr(1009, 1009, go$toNativeArray("Int32", [-80, 0, -80])), new CaseRange.Ptr(1010, 1010, go$toNativeArray("Int32", [7, 0, 7])), new CaseRange.Ptr(1012, 1012, go$toNativeArray("Int32", [0, -60, 0])), new CaseRange.Ptr(1013, 1013, go$toNativeArray("Int32", [-96, 0, -96])), new CaseRange.Ptr(1015, 1016, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1017, 1017, go$toNativeArray("Int32", [0, -7, 0])), new CaseRange.Ptr(1018, 1019, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1021, 1023, go$toNativeArray("Int32", [0, -130, 0])), new CaseRange.Ptr(1024, 1039, go$toNativeArray("Int32", [0, 80, 0])), new CaseRange.Ptr(1040, 1071, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(1072, 1103, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(1104, 1119, go$toNativeArray("Int32", [-80, 0, -80])), new CaseRange.Ptr(1120, 1153, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1162, 1215, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1216, 1216, go$toNativeArray("Int32", [0, 15, 0])), new CaseRange.Ptr(1217, 1230, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1231, 1231, go$toNativeArray("Int32", [-15, 0, -15])), new CaseRange.Ptr(1232, 1319, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(1329, 1366, go$toNativeArray("Int32", [0, 48, 0])), new CaseRange.Ptr(1377, 1414, go$toNativeArray("Int32", [-48, 0, -48])), new CaseRange.Ptr(4256, 4293, go$toNativeArray("Int32", [0, 7264, 0])), new CaseRange.Ptr(4295, 4295, go$toNativeArray("Int32", [0, 7264, 0])), new CaseRange.Ptr(4301, 4301, go$toNativeArray("Int32", [0, 7264, 0])), new CaseRange.Ptr(7545, 7545, go$toNativeArray("Int32", [35332, 0, 35332])), new CaseRange.Ptr(7549, 7549, go$toNativeArray("Int32", [3814, 0, 3814])), new CaseRange.Ptr(7680, 7829, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(7835, 7835, go$toNativeArray("Int32", [-59, 0, -59])), new CaseRange.Ptr(7838, 7838, go$toNativeArray("Int32", [0, -7615, 0])), new CaseRange.Ptr(7840, 7935, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(7936, 7943, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(7944, 7951, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(7952, 7957, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(7960, 7965, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(7968, 7975, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(7976, 7983, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(7984, 7991, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(7992, 7999, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8000, 8005, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8008, 8013, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8017, 8017, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8019, 8019, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8021, 8021, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8023, 8023, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8025, 8025, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8027, 8027, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8029, 8029, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8031, 8031, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8032, 8039, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8040, 8047, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8048, 8049, go$toNativeArray("Int32", [74, 0, 74])), new CaseRange.Ptr(8050, 8053, go$toNativeArray("Int32", [86, 0, 86])), new CaseRange.Ptr(8054, 8055, go$toNativeArray("Int32", [100, 0, 100])), new CaseRange.Ptr(8056, 8057, go$toNativeArray("Int32", [128, 0, 128])), new CaseRange.Ptr(8058, 8059, go$toNativeArray("Int32", [112, 0, 112])), new CaseRange.Ptr(8060, 8061, go$toNativeArray("Int32", [126, 0, 126])), new CaseRange.Ptr(8064, 8071, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8072, 8079, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8080, 8087, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8088, 8095, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8096, 8103, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8104, 8111, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8112, 8113, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8115, 8115, go$toNativeArray("Int32", [9, 0, 9])), new CaseRange.Ptr(8120, 8121, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8122, 8123, go$toNativeArray("Int32", [0, -74, 0])), new CaseRange.Ptr(8124, 8124, go$toNativeArray("Int32", [0, -9, 0])), new CaseRange.Ptr(8126, 8126, go$toNativeArray("Int32", [-7205, 0, -7205])), new CaseRange.Ptr(8131, 8131, go$toNativeArray("Int32", [9, 0, 9])), new CaseRange.Ptr(8136, 8139, go$toNativeArray("Int32", [0, -86, 0])), new CaseRange.Ptr(8140, 8140, go$toNativeArray("Int32", [0, -9, 0])), new CaseRange.Ptr(8144, 8145, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8152, 8153, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8154, 8155, go$toNativeArray("Int32", [0, -100, 0])), new CaseRange.Ptr(8160, 8161, go$toNativeArray("Int32", [8, 0, 8])), new CaseRange.Ptr(8165, 8165, go$toNativeArray("Int32", [7, 0, 7])), new CaseRange.Ptr(8168, 8169, go$toNativeArray("Int32", [0, -8, 0])), new CaseRange.Ptr(8170, 8171, go$toNativeArray("Int32", [0, -112, 0])), new CaseRange.Ptr(8172, 8172, go$toNativeArray("Int32", [0, -7, 0])), new CaseRange.Ptr(8179, 8179, go$toNativeArray("Int32", [9, 0, 9])), new CaseRange.Ptr(8184, 8185, go$toNativeArray("Int32", [0, -128, 0])), new CaseRange.Ptr(8186, 8187, go$toNativeArray("Int32", [0, -126, 0])), new CaseRange.Ptr(8188, 8188, go$toNativeArray("Int32", [0, -9, 0])), new CaseRange.Ptr(8486, 8486, go$toNativeArray("Int32", [0, -7517, 0])), new CaseRange.Ptr(8490, 8490, go$toNativeArray("Int32", [0, -8383, 0])), new CaseRange.Ptr(8491, 8491, go$toNativeArray("Int32", [0, -8262, 0])), new CaseRange.Ptr(8498, 8498, go$toNativeArray("Int32", [0, 28, 0])), new CaseRange.Ptr(8526, 8526, go$toNativeArray("Int32", [-28, 0, -28])), new CaseRange.Ptr(8544, 8559, go$toNativeArray("Int32", [0, 16, 0])), new CaseRange.Ptr(8560, 8575, go$toNativeArray("Int32", [-16, 0, -16])), new CaseRange.Ptr(8579, 8580, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(9398, 9423, go$toNativeArray("Int32", [0, 26, 0])), new CaseRange.Ptr(9424, 9449, go$toNativeArray("Int32", [-26, 0, -26])), new CaseRange.Ptr(11264, 11310, go$toNativeArray("Int32", [0, 48, 0])), new CaseRange.Ptr(11312, 11358, go$toNativeArray("Int32", [-48, 0, -48])), new CaseRange.Ptr(11360, 11361, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11362, 11362, go$toNativeArray("Int32", [0, -10743, 0])), new CaseRange.Ptr(11363, 11363, go$toNativeArray("Int32", [0, -3814, 0])), new CaseRange.Ptr(11364, 11364, go$toNativeArray("Int32", [0, -10727, 0])), new CaseRange.Ptr(11365, 11365, go$toNativeArray("Int32", [-10795, 0, -10795])), new CaseRange.Ptr(11366, 11366, go$toNativeArray("Int32", [-10792, 0, -10792])), new CaseRange.Ptr(11367, 11372, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11373, 11373, go$toNativeArray("Int32", [0, -10780, 0])), new CaseRange.Ptr(11374, 11374, go$toNativeArray("Int32", [0, -10749, 0])), new CaseRange.Ptr(11375, 11375, go$toNativeArray("Int32", [0, -10783, 0])), new CaseRange.Ptr(11376, 11376, go$toNativeArray("Int32", [0, -10782, 0])), new CaseRange.Ptr(11378, 11379, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11381, 11382, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11390, 11391, go$toNativeArray("Int32", [0, -10815, 0])), new CaseRange.Ptr(11392, 11491, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11499, 11502, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11506, 11507, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(11520, 11557, go$toNativeArray("Int32", [-7264, 0, -7264])), new CaseRange.Ptr(11559, 11559, go$toNativeArray("Int32", [-7264, 0, -7264])), new CaseRange.Ptr(11565, 11565, go$toNativeArray("Int32", [-7264, 0, -7264])), new CaseRange.Ptr(42560, 42605, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42624, 42647, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42786, 42799, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42802, 42863, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42873, 42876, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42877, 42877, go$toNativeArray("Int32", [0, -35332, 0])), new CaseRange.Ptr(42878, 42887, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42891, 42892, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42893, 42893, go$toNativeArray("Int32", [0, -42280, 0])), new CaseRange.Ptr(42896, 42899, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42912, 42921, go$toNativeArray("Int32", [1114112, 1114112, 1114112])), new CaseRange.Ptr(42922, 42922, go$toNativeArray("Int32", [0, -42308, 0])), new CaseRange.Ptr(65313, 65338, go$toNativeArray("Int32", [0, 32, 0])), new CaseRange.Ptr(65345, 65370, go$toNativeArray("Int32", [-32, 0, -32])), new CaseRange.Ptr(66560, 66599, go$toNativeArray("Int32", [0, 40, 0])), new CaseRange.Ptr(66600, 66639, go$toNativeArray("Int32", [-40, 0, -40]))]);
		go$pkg.CaseRanges = _CaseRanges;
	}
	return go$pkg;
})();
go$packages["unicode/utf8"] = (function() {
	var go$pkg = {}, decodeRuneInStringInternal, DecodeRuneInString, DecodeLastRuneInString, RuneLen, EncodeRune, RuneCountInString, RuneStart;
	decodeRuneInStringInternal = function(s) {
		var r, size, short$1, n, _tmp, _tmp$1, _tmp$2, c0, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, _tmp$8, _tmp$9, _tmp$10, _tmp$11, c1, _tmp$12, _tmp$13, _tmp$14, _tmp$15, _tmp$16, _tmp$17, _tmp$18, _tmp$19, _tmp$20, _tmp$21, _tmp$22, _tmp$23, c2, _tmp$24, _tmp$25, _tmp$26, _tmp$27, _tmp$28, _tmp$29, _tmp$30, _tmp$31, _tmp$32, _tmp$33, _tmp$34, _tmp$35, _tmp$36, _tmp$37, _tmp$38, c3, _tmp$39, _tmp$40, _tmp$41, _tmp$42, _tmp$43, _tmp$44, _tmp$45, _tmp$46, _tmp$47, _tmp$48, _tmp$49, _tmp$50;
		r = 0;
		size = 0;
		short$1 = false;
		n = s.length;
		if (n < 1) {
			_tmp = 65533; _tmp$1 = 0; _tmp$2 = true; r = _tmp; size = _tmp$1; short$1 = _tmp$2;
			return [r, size, short$1];
		}
		c0 = s.charCodeAt(0);
		if (c0 < 128) {
			_tmp$3 = (c0 >> 0); _tmp$4 = 1; _tmp$5 = false; r = _tmp$3; size = _tmp$4; short$1 = _tmp$5;
			return [r, size, short$1];
		}
		if (c0 < 192) {
			_tmp$6 = 65533; _tmp$7 = 1; _tmp$8 = false; r = _tmp$6; size = _tmp$7; short$1 = _tmp$8;
			return [r, size, short$1];
		}
		if (n < 2) {
			_tmp$9 = 65533; _tmp$10 = 1; _tmp$11 = true; r = _tmp$9; size = _tmp$10; short$1 = _tmp$11;
			return [r, size, short$1];
		}
		c1 = s.charCodeAt(1);
		if (c1 < 128 || 192 <= c1) {
			_tmp$12 = 65533; _tmp$13 = 1; _tmp$14 = false; r = _tmp$12; size = _tmp$13; short$1 = _tmp$14;
			return [r, size, short$1];
		}
		if (c0 < 224) {
			r = ((((c0 & 31) >>> 0) >> 0) << 6 >> 0) | (((c1 & 63) >>> 0) >> 0);
			if (r <= 127) {
				_tmp$15 = 65533; _tmp$16 = 1; _tmp$17 = false; r = _tmp$15; size = _tmp$16; short$1 = _tmp$17;
				return [r, size, short$1];
			}
			_tmp$18 = r; _tmp$19 = 2; _tmp$20 = false; r = _tmp$18; size = _tmp$19; short$1 = _tmp$20;
			return [r, size, short$1];
		}
		if (n < 3) {
			_tmp$21 = 65533; _tmp$22 = 1; _tmp$23 = true; r = _tmp$21; size = _tmp$22; short$1 = _tmp$23;
			return [r, size, short$1];
		}
		c2 = s.charCodeAt(2);
		if (c2 < 128 || 192 <= c2) {
			_tmp$24 = 65533; _tmp$25 = 1; _tmp$26 = false; r = _tmp$24; size = _tmp$25; short$1 = _tmp$26;
			return [r, size, short$1];
		}
		if (c0 < 240) {
			r = (((((c0 & 15) >>> 0) >> 0) << 12 >> 0) | ((((c1 & 63) >>> 0) >> 0) << 6 >> 0)) | (((c2 & 63) >>> 0) >> 0);
			if (r <= 2047) {
				_tmp$27 = 65533; _tmp$28 = 1; _tmp$29 = false; r = _tmp$27; size = _tmp$28; short$1 = _tmp$29;
				return [r, size, short$1];
			}
			if (55296 <= r && r <= 57343) {
				_tmp$30 = 65533; _tmp$31 = 1; _tmp$32 = false; r = _tmp$30; size = _tmp$31; short$1 = _tmp$32;
				return [r, size, short$1];
			}
			_tmp$33 = r; _tmp$34 = 3; _tmp$35 = false; r = _tmp$33; size = _tmp$34; short$1 = _tmp$35;
			return [r, size, short$1];
		}
		if (n < 4) {
			_tmp$36 = 65533; _tmp$37 = 1; _tmp$38 = true; r = _tmp$36; size = _tmp$37; short$1 = _tmp$38;
			return [r, size, short$1];
		}
		c3 = s.charCodeAt(3);
		if (c3 < 128 || 192 <= c3) {
			_tmp$39 = 65533; _tmp$40 = 1; _tmp$41 = false; r = _tmp$39; size = _tmp$40; short$1 = _tmp$41;
			return [r, size, short$1];
		}
		if (c0 < 248) {
			r = ((((((c0 & 7) >>> 0) >> 0) << 18 >> 0) | ((((c1 & 63) >>> 0) >> 0) << 12 >> 0)) | ((((c2 & 63) >>> 0) >> 0) << 6 >> 0)) | (((c3 & 63) >>> 0) >> 0);
			if (r <= 65535 || 1114111 < r) {
				_tmp$42 = 65533; _tmp$43 = 1; _tmp$44 = false; r = _tmp$42; size = _tmp$43; short$1 = _tmp$44;
				return [r, size, short$1];
			}
			_tmp$45 = r; _tmp$46 = 4; _tmp$47 = false; r = _tmp$45; size = _tmp$46; short$1 = _tmp$47;
			return [r, size, short$1];
		}
		_tmp$48 = 65533; _tmp$49 = 1; _tmp$50 = false; r = _tmp$48; size = _tmp$49; short$1 = _tmp$50;
		return [r, size, short$1];
	};
	DecodeRuneInString = go$pkg.DecodeRuneInString = function(s) {
		var r, size, _tuple;
		r = 0;
		size = 0;
		_tuple = decodeRuneInStringInternal(s); r = _tuple[0]; size = _tuple[1];
		return [r, size];
	};
	DecodeLastRuneInString = go$pkg.DecodeLastRuneInString = function(s) {
		var r, size, end, _tmp, _tmp$1, start, _tmp$2, _tmp$3, lim, _tuple, _tmp$4, _tmp$5, _tmp$6, _tmp$7;
		r = 0;
		size = 0;
		end = s.length;
		if (end === 0) {
			_tmp = 65533; _tmp$1 = 0; r = _tmp; size = _tmp$1;
			return [r, size];
		}
		start = end - 1 >> 0;
		r = (s.charCodeAt(start) >> 0);
		if (r < 128) {
			_tmp$2 = r; _tmp$3 = 1; r = _tmp$2; size = _tmp$3;
			return [r, size];
		}
		lim = end - 4 >> 0;
		if (lim < 0) {
			lim = 0;
		}
		start = start - 1 >> 0;
		while (start >= lim) {
			if (RuneStart(s.charCodeAt(start))) {
				break;
			}
			start = start - 1 >> 0;
		}
		if (start < 0) {
			start = 0;
		}
		_tuple = DecodeRuneInString(s.substring(start, end)); r = _tuple[0]; size = _tuple[1];
		if (!(((start + size >> 0) === end))) {
			_tmp$4 = 65533; _tmp$5 = 1; r = _tmp$4; size = _tmp$5;
			return [r, size];
		}
		_tmp$6 = r; _tmp$7 = size; r = _tmp$6; size = _tmp$7;
		return [r, size];
	};
	RuneLen = go$pkg.RuneLen = function(r) {
		if (r < 0) {
			return -1;
		} else if (r <= 127) {
			return 1;
		} else if (r <= 2047) {
			return 2;
		} else if (55296 <= r && r <= 57343) {
			return -1;
		} else if (r <= 65535) {
			return 3;
		} else if (r <= 1114111) {
			return 4;
		}
		return -1;
	};
	EncodeRune = go$pkg.EncodeRune = function(p, r) {
		if ((r >>> 0) <= 127) {
			(0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0] = (r << 24 >>> 24);
			return 1;
		}
		if ((r >>> 0) <= 2047) {
			(0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0] = (192 | ((r >> 6 >> 0) << 24 >>> 24)) >>> 0;
			(1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1] = (128 | (((r << 24 >>> 24) & 63) >>> 0)) >>> 0;
			return 2;
		}
		if ((r >>> 0) > 1114111) {
			r = 65533;
		}
		if (55296 <= r && r <= 57343) {
			r = 65533;
		}
		if ((r >>> 0) <= 65535) {
			(0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0] = (224 | ((r >> 12 >> 0) << 24 >>> 24)) >>> 0;
			(1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1] = (128 | ((((r >> 6 >> 0) << 24 >>> 24) & 63) >>> 0)) >>> 0;
			(2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2] = (128 | (((r << 24 >>> 24) & 63) >>> 0)) >>> 0;
			return 3;
		}
		(0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0] = (240 | ((r >> 18 >> 0) << 24 >>> 24)) >>> 0;
		(1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1] = (128 | ((((r >> 12 >> 0) << 24 >>> 24) & 63) >>> 0)) >>> 0;
		(2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2] = (128 | ((((r >> 6 >> 0) << 24 >>> 24) & 63) >>> 0)) >>> 0;
		(3 < 0 || 3 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 3] = (128 | (((r << 24 >>> 24) & 63) >>> 0)) >>> 0;
		return 4;
	};
	RuneCountInString = go$pkg.RuneCountInString = function(s) {
		var n, _ref, _i, _rune;
		n = 0;
		_ref = s;
		_i = 0;
		while (_i < _ref.length) {
			_rune = go$decodeRune(_ref, _i);
			n = n + 1 >> 0;
			_i += _rune[1];
		}
		return n;
	};
	RuneStart = go$pkg.RuneStart = function(b) {
		return !((((b & 192) >>> 0) === 128));
	};
	go$pkg.init = function() {
	}
	return go$pkg;
})();
go$packages["strings"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], errors = go$packages["errors"], io = go$packages["io"], utf8 = go$packages["unicode/utf8"], unicode = go$packages["unicode"], IndexByte, hashstr, Index, IndexRune, Join, Map, ToLower, TrimLeftFunc, TrimRightFunc, TrimFunc, indexFunc, lastIndexFunc, makeCutsetFunc, Trim;
	IndexByte = go$pkg.IndexByte = function(s, c) {
		return go$parseInt(s.indexOf(go$global.String.fromCharCode(c))) >> 0;
	};
	hashstr = function(sep) {
		var hash, i, _tmp, _tmp$1, pow, sq, i$1, x, x$1;
		hash = 0;
		i = 0;
		while (i < sep.length) {
			hash = ((((hash >>> 16 << 16) * 16777619 >>> 0) + (hash << 16 >>> 16) * 16777619) >>> 0) + (sep.charCodeAt(i) >>> 0) >>> 0;
			i = i + 1 >> 0;
		}
		_tmp = 1; _tmp$1 = 16777619; pow = _tmp; sq = _tmp$1;
		i$1 = sep.length;
		while (i$1 > 0) {
			if (!(((i$1 & 1) === 0))) {
				pow = (x = sq, (((pow >>> 16 << 16) * x >>> 0) + (pow << 16 >>> 16) * x) >>> 0);
			}
			sq = (x$1 = sq, (((sq >>> 16 << 16) * x$1 >>> 0) + (sq << 16 >>> 16) * x$1) >>> 0);
			i$1 = i$1 >> 1 >> 0;
		}
		return [hash, pow];
	};
	Index = go$pkg.Index = function(s, sep) {
		var n, _tuple, hashsep, pow, h, i, i$1, x;
		n = sep.length;
		if (n === 0) {
			return 0;
		} else if (n === 1) {
			return IndexByte(s, sep.charCodeAt(0));
		} else if (n === s.length) {
			if (sep === s) {
				return 0;
			}
			return -1;
		} else if (n > s.length) {
			return -1;
		}
		_tuple = hashstr(sep); hashsep = _tuple[0]; pow = _tuple[1];
		h = 0;
		i = 0;
		while (i < n) {
			h = ((((h >>> 16 << 16) * 16777619 >>> 0) + (h << 16 >>> 16) * 16777619) >>> 0) + (s.charCodeAt(i) >>> 0) >>> 0;
			i = i + 1 >> 0;
		}
		if ((h === hashsep) && s.substring(0, n) === sep) {
			return 0;
		}
		i$1 = n;
		while (i$1 < s.length) {
			h = (((h >>> 16 << 16) * 16777619 >>> 0) + (h << 16 >>> 16) * 16777619) >>> 0;
			h = h + ((s.charCodeAt(i$1) >>> 0)) >>> 0;
			h = h - ((x = (s.charCodeAt((i$1 - n >> 0)) >>> 0), (((pow >>> 16 << 16) * x >>> 0) + (pow << 16 >>> 16) * x) >>> 0)) >>> 0;
			i$1 = i$1 + 1 >> 0;
			if ((h === hashsep) && s.substring((i$1 - n >> 0), i$1) === sep) {
				return i$1 - n >> 0;
			}
		}
		return -1;
	};
	IndexRune = go$pkg.IndexRune = function(s, r) {
		var b, i, _ref, _i, _rune, c, i$1;
		if (r < 128) {
			b = (r << 24 >>> 24);
			i = 0;
			while (i < s.length) {
				if (s.charCodeAt(i) === b) {
					return i;
				}
				i = i + 1 >> 0;
			}
		} else {
			_ref = s;
			_i = 0;
			while (_i < _ref.length) {
				_rune = go$decodeRune(_ref, _i);
				c = _rune[0];
				i$1 = _i;
				if (c === r) {
					return i$1;
				}
				_i += _rune[1];
			}
		}
		return -1;
	};
	Join = go$pkg.Join = function(a, sep) {
		var x, x$1, n, i, b, bp, _ref, _i, s;
		if (a.length === 0) {
			return "";
		}
		if (a.length === 1) {
			return ((0 < 0 || 0 >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + 0]);
		}
		n = (x = sep.length, x$1 = (a.length - 1 >> 0), (((x >>> 16 << 16) * x$1 >> 0) + (x << 16 >>> 16) * x$1) >> 0);
		i = 0;
		while (i < a.length) {
			n = n + (((i < 0 || i >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + i]).length) >> 0;
			i = i + 1 >> 0;
		}
		b = (go$sliceType(Go$Uint8)).make(n, 0, function() { return 0; });
		bp = go$copyString(b, ((0 < 0 || 0 >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + 0]));
		_ref = go$subslice(a, 1);
		_i = 0;
		while (_i < _ref.length) {
			s = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			bp = bp + (go$copyString(go$subslice(b, bp), sep)) >> 0;
			bp = bp + (go$copyString(go$subslice(b, bp), s)) >> 0;
			_i++;
		}
		return go$bytesToString(b);
	};
	Map = go$pkg.Map = function(mapping, s) {
		var maxbytes, nbytes, b, _ref, _i, _rune, c, i, r, wid, nb;
		maxbytes = s.length;
		nbytes = 0;
		b = (go$sliceType(Go$Uint8)).nil;
		_ref = s;
		_i = 0;
		while (_i < _ref.length) {
			_rune = go$decodeRune(_ref, _i);
			c = _rune[0];
			i = _i;
			r = mapping(c);
			if (b === (go$sliceType(Go$Uint8)).nil) {
				if (r === c) {
					_i += _rune[1];
					continue;
				}
				b = (go$sliceType(Go$Uint8)).make(maxbytes, 0, function() { return 0; });
				nbytes = go$copyString(b, s.substring(0, i));
			}
			if (r >= 0) {
				wid = 1;
				if (r >= 128) {
					wid = utf8.RuneLen(r);
				}
				if ((nbytes + wid >> 0) > maxbytes) {
					maxbytes = ((((maxbytes >>> 16 << 16) * 2 >> 0) + (maxbytes << 16 >>> 16) * 2) >> 0) + 4 >> 0;
					nb = (go$sliceType(Go$Uint8)).make(maxbytes, 0, function() { return 0; });
					go$copySlice(nb, go$subslice(b, 0, nbytes));
					b = nb;
				}
				nbytes = nbytes + (utf8.EncodeRune(go$subslice(b, nbytes, maxbytes), r)) >> 0;
			}
			_i += _rune[1];
		}
		if (b === (go$sliceType(Go$Uint8)).nil) {
			return s;
		}
		return go$bytesToString(go$subslice(b, 0, nbytes));
	};
	ToLower = go$pkg.ToLower = function(s) {
		return Map(unicode.ToLower, s);
	};
	TrimLeftFunc = go$pkg.TrimLeftFunc = function(s, f) {
		var i;
		i = indexFunc(s, f, false);
		if (i === -1) {
			return "";
		}
		return s.substring(i);
	};
	TrimRightFunc = go$pkg.TrimRightFunc = function(s, f) {
		var i, _tuple, wid;
		i = lastIndexFunc(s, f, false);
		if (i >= 0 && s.charCodeAt(i) >= 128) {
			_tuple = utf8.DecodeRuneInString(s.substring(i)); wid = _tuple[1];
			i = i + (wid) >> 0;
		} else {
			i = i + 1 >> 0;
		}
		return s.substring(0, i);
	};
	TrimFunc = go$pkg.TrimFunc = function(s, f) {
		return TrimRightFunc(TrimLeftFunc(s, f), f);
	};
	indexFunc = function(s, f, truth) {
		var start, wid, r, _tuple;
		start = 0;
		while (start < s.length) {
			wid = 1;
			r = (s.charCodeAt(start) >> 0);
			if (r >= 128) {
				_tuple = utf8.DecodeRuneInString(s.substring(start)); r = _tuple[0]; wid = _tuple[1];
			}
			if (f(r) === truth) {
				return start;
			}
			start = start + (wid) >> 0;
		}
		return -1;
	};
	lastIndexFunc = function(s, f, truth) {
		var i, _tuple, r, size;
		i = s.length;
		while (i > 0) {
			_tuple = utf8.DecodeLastRuneInString(s.substring(0, i)); r = _tuple[0]; size = _tuple[1];
			i = i - (size) >> 0;
			if (f(r) === truth) {
				return i;
			}
		}
		return -1;
	};
	makeCutsetFunc = function(cutset) {
		return (function(r) {
			return IndexRune(cutset, r) >= 0;
		});
	};
	Trim = go$pkg.Trim = function(s, cutset) {
		if (s === "" || cutset === "") {
			return s;
		}
		return TrimFunc(s, makeCutsetFunc(cutset));
	};
	go$pkg.init = function() {
	}
	return go$pkg;
})();
go$packages["math"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], Ldexp, Float32bits, Float32frombits, Float64bits, math, zero, negInf, nan, pow10tab;
	Ldexp = go$pkg.Ldexp = function(frac, exp$1) {
		if (frac === 0) {
			return frac;
		}
		if (exp$1 >= 1024) {
			return frac * go$parseFloat(math.pow(2, 1023)) * go$parseFloat(math.pow(2, exp$1 - 1023 >> 0));
		}
		if (exp$1 <= -1024) {
			return frac * go$parseFloat(math.pow(2, -1023)) * go$parseFloat(math.pow(2, exp$1 + 1023 >> 0));
		}
		return frac * go$parseFloat(math.pow(2, exp$1));
	};
	Float32bits = go$pkg.Float32bits = function(f) {
		var s, e, r;
		if (go$float32IsEqual(f, 0)) {
			if (go$float32IsEqual(1 / f, negInf)) {
				return 2147483648;
			}
			return 0;
		}
		if (!((go$float32IsEqual(f, f)))) {
			return 2143289344;
		}
		s = 0;
		if (f < 0) {
			s = 2147483648;
			f = -f;
		}
		e = 150;
		while (f >= 1.6777216e+07) {
			f = f / 2;
			if (e === 255) {
				break;
			}
			e = e + 1 >>> 0;
		}
		while (f < 8.388608e+06) {
			e = e - 1 >>> 0;
			if (e === 0) {
				break;
			}
			f = f * 2;
		}
		r = go$parseFloat(go$mod(f, 2));
		if ((r > 0.5 && r < 1) || r >= 1.5) {
			f = f + 1;
		}
		return (((s | (e << 23 >>> 0)) >>> 0) | (((f >> 0) & ~8388608))) >>> 0;
	};
	Float32frombits = go$pkg.Float32frombits = function(b) {
		var s, e, m;
		s = 1;
		if (!((((b & 2147483648) >>> 0) === 0))) {
			s = -1;
		}
		e = (((b >>> 23 >>> 0)) & 255) >>> 0;
		m = (b & 8388607) >>> 0;
		if (e === 255) {
			if (m === 0) {
				return s / 0;
			}
			return nan;
		}
		if (!((e === 0))) {
			m = m + 8388608 >>> 0;
		}
		if (e === 0) {
			e = 1;
		}
		return Ldexp(m, ((e >> 0) - 127 >> 0) - 23 >> 0) * s;
	};
	Float64bits = go$pkg.Float64bits = function(f) {
		var s, e, x, x$1, x$2, x$3;
		if (f === 0) {
			if (1 / f === negInf) {
				return new Go$Uint64(2147483648, 0);
			}
			return new Go$Uint64(0, 0);
		}
		if (!((f === f))) {
			return new Go$Uint64(2146959360, 1);
		}
		s = new Go$Uint64(0, 0);
		if (f < 0) {
			s = new Go$Uint64(2147483648, 0);
			f = -f;
		}
		e = 1075;
		while (f >= 9.007199254740992e+15) {
			f = f / 2;
			if (e === 2047) {
				break;
			}
			e = e + 1 >>> 0;
		}
		while (f < 4.503599627370496e+15) {
			e = e - 1 >>> 0;
			if (e === 0) {
				break;
			}
			f = f * 2;
		}
		return (x = (x$1 = go$shiftLeft64(new Go$Uint64(0, e), 52), new Go$Uint64(s.high | x$1.high, (s.low | x$1.low) >>> 0)), x$2 = (x$3 = new Go$Uint64(0, f), new Go$Uint64(x$3.high &~ 1048576, (x$3.low &~ 0) >>> 0)), new Go$Uint64(x.high | x$2.high, (x.low | x$2.low) >>> 0));
	};
	go$pkg.init = function() {
		pow10tab = go$makeNativeArray("Float64", 70, function() { return 0; });
		math = go$global.Math;
		zero = 0;
		negInf = -1 / zero;
		nan = 0 / zero;
		var i, _q, m;
		Float32bits(0);
		Float32frombits(0);
		pow10tab[0] = 1;
		pow10tab[1] = 10;
		i = 2;
		while (i < 70) {
			m = (_q = i / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
			pow10tab[i] = pow10tab[m] * pow10tab[(i - m >> 0)];
			i = i + 1 >> 0;
		}
	}
	return go$pkg;
})();
go$packages["syscall"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], sync = go$packages["sync"], runtime = go$packages["runtime"], errors$1 = go$packages["errors"], mmapper, Errno, Timespec, Stat_t, Dirent, syscall, Syscall, Syscall6, BytePtrFromString, copyenv, Getenv, itoa, ReadDirent, ParseDirent, Read, Write, Close, Fchdir, Fchmod, Fchown, Fstat, Fsync, Ftruncate, Getdirentries, Lstat, Open, Pread, Pwrite, read, Seek, write, mmap, munmap, syscallModule, envOnce, envLock, env, envs, mapper, errors;
	mmapper = go$pkg.mmapper = go$newType(0, "Struct", "syscall.mmapper", "mmapper", "syscall", function(Mutex_, active_, mmap_, munmap_) {
		this.go$val = this;
		this.Mutex = Mutex_ !== undefined ? Mutex_ : new sync.Mutex.Ptr();
		this.active = active_ !== undefined ? active_ : false;
		this.mmap = mmap_ !== undefined ? mmap_ : go$throwNilPointerError;
		this.munmap = munmap_ !== undefined ? munmap_ : go$throwNilPointerError;
	});
	Errno = go$pkg.Errno = go$newType(4, "Uintptr", "syscall.Errno", "Errno", "syscall", null);
	Timespec = go$pkg.Timespec = go$newType(0, "Struct", "syscall.Timespec", "Timespec", "syscall", function(Sec_, Nsec_) {
		this.go$val = this;
		this.Sec = Sec_ !== undefined ? Sec_ : new Go$Int64(0, 0);
		this.Nsec = Nsec_ !== undefined ? Nsec_ : new Go$Int64(0, 0);
	});
	Stat_t = go$pkg.Stat_t = go$newType(0, "Struct", "syscall.Stat_t", "Stat_t", "syscall", function(Dev_, Mode_, Nlink_, Ino_, Uid_, Gid_, Rdev_, Pad_cgo_0_, Atimespec_, Mtimespec_, Ctimespec_, Birthtimespec_, Size_, Blocks_, Blksize_, Flags_, Gen_, Lspare_, Qspare_) {
		this.go$val = this;
		this.Dev = Dev_ !== undefined ? Dev_ : 0;
		this.Mode = Mode_ !== undefined ? Mode_ : 0;
		this.Nlink = Nlink_ !== undefined ? Nlink_ : 0;
		this.Ino = Ino_ !== undefined ? Ino_ : new Go$Uint64(0, 0);
		this.Uid = Uid_ !== undefined ? Uid_ : 0;
		this.Gid = Gid_ !== undefined ? Gid_ : 0;
		this.Rdev = Rdev_ !== undefined ? Rdev_ : 0;
		this.Pad_cgo_0 = Pad_cgo_0_ !== undefined ? Pad_cgo_0_ : go$makeNativeArray("Uint8", 4, function() { return 0; });
		this.Atimespec = Atimespec_ !== undefined ? Atimespec_ : new Timespec.Ptr();
		this.Mtimespec = Mtimespec_ !== undefined ? Mtimespec_ : new Timespec.Ptr();
		this.Ctimespec = Ctimespec_ !== undefined ? Ctimespec_ : new Timespec.Ptr();
		this.Birthtimespec = Birthtimespec_ !== undefined ? Birthtimespec_ : new Timespec.Ptr();
		this.Size = Size_ !== undefined ? Size_ : new Go$Int64(0, 0);
		this.Blocks = Blocks_ !== undefined ? Blocks_ : new Go$Int64(0, 0);
		this.Blksize = Blksize_ !== undefined ? Blksize_ : 0;
		this.Flags = Flags_ !== undefined ? Flags_ : 0;
		this.Gen = Gen_ !== undefined ? Gen_ : 0;
		this.Lspare = Lspare_ !== undefined ? Lspare_ : 0;
		this.Qspare = Qspare_ !== undefined ? Qspare_ : go$makeNativeArray("Int64", 2, function() { return new Go$Int64(0, 0); });
	});
	Dirent = go$pkg.Dirent = go$newType(0, "Struct", "syscall.Dirent", "Dirent", "syscall", function(Ino_, Seekoff_, Reclen_, Namlen_, Type_, Name_, Pad_cgo_0_) {
		this.go$val = this;
		this.Ino = Ino_ !== undefined ? Ino_ : new Go$Uint64(0, 0);
		this.Seekoff = Seekoff_ !== undefined ? Seekoff_ : new Go$Uint64(0, 0);
		this.Reclen = Reclen_ !== undefined ? Reclen_ : 0;
		this.Namlen = Namlen_ !== undefined ? Namlen_ : 0;
		this.Type = Type_ !== undefined ? Type_ : 0;
		this.Name = Name_ !== undefined ? Name_ : go$makeNativeArray("Int8", 1024, function() { return 0; });
		this.Pad_cgo_0 = Pad_cgo_0_ !== undefined ? Pad_cgo_0_ : go$makeNativeArray("Uint8", 3, function() { return 0; });
	});
	syscall = function(name) {
		var require, syscallHandler;
		if (go$interfaceIsEqual(syscallModule, null)) {
			require = go$global.require;
			if (require === undefined) {
				syscallHandler = go$global.$syscall;
				if (!(syscallHandler === undefined)) {
					return syscallHandler;
				}
				throw go$panic(new Go$String("syscalls not available"));
			}
			syscallModule = require(go$externalize("syscall", Go$String));
		}
		return syscallModule[go$externalize(name, Go$String)];
	};
	Syscall = go$pkg.Syscall = function(trap, a1, a2, a3) {
		var r1, r2, err, r, _tmp, _tmp$1, _tmp$2;
		r1 = 0;
		r2 = 0;
		err = 0;
		r = syscall("Syscall")(trap, a1, a2, a3);
		_tmp = ((go$parseInt(r[0]) >> 0) >>> 0); _tmp$1 = ((go$parseInt(r[1]) >> 0) >>> 0); _tmp$2 = ((go$parseInt(r[2]) >> 0) >>> 0); r1 = _tmp; r2 = _tmp$1; err = _tmp$2;
		return [r1, r2, err];
	};
	Syscall6 = go$pkg.Syscall6 = function(trap, a1, a2, a3, a4, a5, a6) {
		var r1, r2, err, r, _tmp, _tmp$1, _tmp$2;
		r1 = 0;
		r2 = 0;
		err = 0;
		r = syscall("Syscall6")(trap, a1, a2, a3, a4, a5, a6);
		_tmp = ((go$parseInt(r[0]) >> 0) >>> 0); _tmp$1 = ((go$parseInt(r[1]) >> 0) >>> 0); _tmp$2 = ((go$parseInt(r[2]) >> 0) >>> 0); r1 = _tmp; r2 = _tmp$1; err = _tmp$2;
		return [r1, r2, err];
	};
	BytePtrFromString = go$pkg.BytePtrFromString = function(s) {
		return [go$stringToBytes(go$externalize(s, Go$String), go$externalize(true, Go$Bool)), null];
	};
	copyenv = function() {
		var _ref, _i, s, i, j, key, _tuple, _entry, ok, _key;
		env = new Go$Map();
		_ref = envs;
		_i = 0;
		while (_i < _ref.length) {
			s = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			j = 0;
			while (j < s.length) {
				if (s.charCodeAt(j) === 61) {
					key = s.substring(0, j);
					_tuple = (_entry = env[key], _entry !== undefined ? [_entry.v, true] : [0, false]); ok = _tuple[1];
					if (!ok) {
						_key = key; (env || go$throwRuntimeError("assignment to entry in nil map"))[_key] = { k: _key, v: i };
					}
					break;
				}
				j = j + 1 >> 0;
			}
			_i++;
		}
	};
	Getenv = go$pkg.Getenv = function(key) {
		var value, found, _tmp, _tmp$1, _tuple, _entry, i, ok, _tmp$2, _tmp$3, s, i$1, _tmp$4, _tmp$5, _tmp$6, _tmp$7;
		value = "";
		found = false;
		var go$deferred = [];
		try {
			envOnce.Do(copyenv);
			if (key.length === 0) {
				_tmp = ""; _tmp$1 = false; value = _tmp; found = _tmp$1;
				return [value, found];
			}
			envLock.RLock();
			go$deferred.push({ recv: envLock, method: "RUnlock", args: [] });
			_tuple = (_entry = env[key], _entry !== undefined ? [_entry.v, true] : [0, false]); i = _tuple[0]; ok = _tuple[1];
			if (!ok) {
				_tmp$2 = ""; _tmp$3 = false; value = _tmp$2; found = _tmp$3;
				return [value, found];
			}
			s = ((i < 0 || i >= envs.length) ? go$throwRuntimeError("index out of range") : envs.array[envs.offset + i]);
			i$1 = 0;
			while (i$1 < s.length) {
				if (s.charCodeAt(i$1) === 61) {
					_tmp$4 = s.substring((i$1 + 1 >> 0)); _tmp$5 = true; value = _tmp$4; found = _tmp$5;
					return [value, found];
				}
				i$1 = i$1 + 1 >> 0;
			}
			_tmp$6 = ""; _tmp$7 = false; value = _tmp$6; found = _tmp$7;
			return [value, found];
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
			return [value, found];
		}
	};
	itoa = function(val) {
		var buf, i, _r, _q;
		if (val < 0) {
			return "-" + itoa(-val);
		}
		buf = go$makeNativeArray("Uint8", 32, function() { return 0; });
		i = 31;
		while (val >= 10) {
			buf[i] = (((_r = val % 10, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) + 48 >> 0) << 24 >>> 24);
			i = i - 1 >> 0;
			val = (_q = val / 10, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		}
		buf[i] = ((val + 48 >> 0) << 24 >>> 24);
		return go$bytesToString(go$subslice(new (go$sliceType(Go$Uint8))(buf), i));
	};
	Timespec.Ptr.prototype.Unix = function() {
		var sec, nsec, ts, _tmp, _tmp$1;
		sec = new Go$Int64(0, 0);
		nsec = new Go$Int64(0, 0);
		ts = this;
		_tmp = ts.Sec; _tmp$1 = ts.Nsec; sec = _tmp; nsec = _tmp$1;
		return [sec, nsec];
	};
	Timespec.prototype.Unix = function() { return this.go$val.Unix(); };
	Timespec.Ptr.prototype.Nano = function() {
		var ts, x, x$1;
		ts = this;
		return (x = go$mul64(ts.Sec, new Go$Int64(0, 1000000000)), x$1 = ts.Nsec, new Go$Int64(x.high + x$1.high, x.low + x$1.low));
	};
	Timespec.prototype.Nano = function() { return this.go$val.Nano(); };
	ReadDirent = go$pkg.ReadDirent = function(fd, buf) {
		var n, err, _tuple;
		n = 0;
		err = null;
		_tuple = Getdirentries(fd, buf, new Uint8Array(8)); n = _tuple[0]; err = _tuple[1];
		return [n, err];
	};
	ParseDirent = go$pkg.ParseDirent = function(buf, max, names) {
		var consumed, count, newnames, origlen, dirent, _array, _struct, _view, x, bytes, name, _tmp, _tmp$1, _tmp$2;
		consumed = 0;
		count = 0;
		newnames = (go$sliceType(Go$String)).nil;
		origlen = buf.length;
		while (!((max === 0)) && buf.length > 0) {
			dirent = [undefined];
			dirent[0] = (_array = go$sliceToArray(buf), _struct = new Dirent.Ptr(), _view = new DataView(_array.buffer, _array.byteOffset), _struct.Ino = new Go$Uint64(_view.getUint32(4, true), _view.getUint32(0, true)), _struct.Seekoff = new Go$Uint64(_view.getUint32(12, true), _view.getUint32(8, true)), _struct.Reclen = _view.getUint16(16, true), _struct.Namlen = _view.getUint16(18, true), _struct.Type = _view.getUint8(20, true), _struct.Name = new (go$nativeArray("Int8"))(_array.buffer, go$min(_array.byteOffset + 21, _array.buffer.byteLength)), _struct.Pad_cgo_0 = new (go$nativeArray("Uint8"))(_array.buffer, go$min(_array.byteOffset + 1045, _array.buffer.byteLength)), _struct);
			if (dirent[0].Reclen === 0) {
				buf = (go$sliceType(Go$Uint8)).nil;
				break;
			}
			buf = go$subslice(buf, dirent[0].Reclen);
			if ((x = dirent[0].Ino, (x.high === 0 && x.low === 0))) {
				continue;
			}
			bytes = go$sliceToArray(new (go$sliceType(Go$Uint8))(dirent[0].Name));
			name = go$bytesToString(go$subslice(new (go$sliceType(Go$Uint8))(bytes), 0, dirent[0].Namlen));
			if (name === "." || name === "..") {
				continue;
			}
			max = max - 1 >> 0;
			count = count + 1 >> 0;
			names = go$append(names, name);
		}
		_tmp = origlen - buf.length >> 0; _tmp$1 = count; _tmp$2 = names; consumed = _tmp; count = _tmp$1; newnames = _tmp$2;
		return [consumed, count, newnames];
	};
	mmapper.Ptr.prototype.Mmap = function(fd, offset, length, prot, flags) {
		var data, err, m, _tmp, _tmp$1, _tuple, addr, errno, _tmp$2, _tmp$3, sl, b, v, x, x$1, p, _key, _tmp$4, _tmp$5;
		data = (go$sliceType(Go$Uint8)).nil;
		err = null;
		var go$deferred = [];
		try {
			m = this;
			if (length <= 0) {
				_tmp = (go$sliceType(Go$Uint8)).nil; _tmp$1 = new Errno(22); data = _tmp; err = _tmp$1;
				return [data, err];
			}
			_tuple = m.mmap(0, (length >>> 0), prot, flags, fd, offset); addr = _tuple[0]; errno = _tuple[1];
			if (!(go$interfaceIsEqual(errno, null))) {
				_tmp$2 = (go$sliceType(Go$Uint8)).nil; _tmp$3 = errno; data = _tmp$2; err = _tmp$3;
				return [data, err];
			}
			sl = new (go$structType([["addr", "addr", "syscall", Go$Uintptr, ""], ["len", "len", "syscall", Go$Int, ""], ["cap", "cap", "syscall", Go$Int, ""]])).Ptr(addr, length, length);
			b = sl;
			p = new (go$ptrType(Go$Uint8))(function() { return (x$1 = b.capacity - 1 >> 0, ((x$1 < 0 || x$1 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + x$1])); }, function(v) { (x = b.capacity - 1 >> 0, (x < 0 || x >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + x] = v);; });
			m.Mutex.Lock();
			go$deferred.push({ recv: m, method: "Unlock", args: [] });
			_key = p; (m.active || go$throwRuntimeError("assignment to entry in nil map"))[_key.go$key()] = { k: _key, v: b };
			_tmp$4 = b; _tmp$5 = null; data = _tmp$4; err = _tmp$5;
			return [data, err];
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
			return [data, err];
		}
	};
	mmapper.prototype.Mmap = function(fd, offset, length, prot, flags) { return this.go$val.Mmap(fd, offset, length, prot, flags); };
	mmapper.Ptr.prototype.Munmap = function(data) {
		var err, m, v, x, x$1, p, _entry, b, errno;
		err = null;
		var go$deferred = [];
		try {
			m = this;
			if ((data.length === 0) || !((data.length === data.capacity))) {
				err = new Errno(22);
				return err;
			}
			p = new (go$ptrType(Go$Uint8))(function() { return (x$1 = data.capacity - 1 >> 0, ((x$1 < 0 || x$1 >= data.length) ? go$throwRuntimeError("index out of range") : data.array[data.offset + x$1])); }, function(v) { (x = data.capacity - 1 >> 0, (x < 0 || x >= data.length) ? go$throwRuntimeError("index out of range") : data.array[data.offset + x] = v);; });
			m.Mutex.Lock();
			go$deferred.push({ recv: m, method: "Unlock", args: [] });
			b = (_entry = m.active[p.go$key()], _entry !== undefined ? _entry.v : (go$sliceType(Go$Uint8)).nil);
			if (b === (go$sliceType(Go$Uint8)).nil || !(go$sliceIsEqual(b, 0, data, 0))) {
				err = new Errno(22);
				return err;
			}
			errno = m.munmap(go$sliceToArray(b), (b.length >>> 0));
			if (!(go$interfaceIsEqual(errno, null))) {
				err = errno;
				return err;
			}
			delete m.active[p.go$key()];
			err = null;
			return err;
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
			return err;
		}
	};
	mmapper.prototype.Munmap = function(data) { return this.go$val.Munmap(data); };
	Errno.prototype.Error = function() {
		var e, s;
		e = this.go$val;
		if (0 <= (e >> 0) && (e >> 0) < 106) {
			s = errors[e];
			if (!(s === "")) {
				return s;
			}
		}
		return "errno " + itoa((e >> 0));
	};
	go$ptrType(Errno).prototype.Error = function() { return new Errno(this.go$get()).Error(); };
	Errno.prototype.Temporary = function() {
		var e;
		e = this.go$val;
		return (e === 4) || (e === 24) || (new Errno(e)).Timeout();
	};
	go$ptrType(Errno).prototype.Temporary = function() { return new Errno(this.go$get()).Temporary(); };
	Errno.prototype.Timeout = function() {
		var e;
		e = this.go$val;
		return (e === 35) || (e === 35) || (e === 60);
	};
	go$ptrType(Errno).prototype.Timeout = function() { return new Errno(this.go$get()).Timeout(); };
	Read = go$pkg.Read = function(fd, p) {
		var n, err, _tuple;
		n = 0;
		err = null;
		_tuple = read(fd, p); n = _tuple[0]; err = _tuple[1];
		return [n, err];
	};
	Write = go$pkg.Write = function(fd, p) {
		var n, err, _tuple;
		n = 0;
		err = null;
		_tuple = write(fd, p); n = _tuple[0]; err = _tuple[1];
		return [n, err];
	};
	Close = go$pkg.Close = function(fd) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(6, (fd >>> 0), 0, 0); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Fchdir = go$pkg.Fchdir = function(fd) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(13, (fd >>> 0), 0, 0); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Fchmod = go$pkg.Fchmod = function(fd, mode) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(124, (fd >>> 0), (mode >>> 0), 0); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Fchown = go$pkg.Fchown = function(fd, uid, gid) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(123, (fd >>> 0), (uid >>> 0), (gid >>> 0)); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Fstat = go$pkg.Fstat = function(fd, stat) {
		var err, _tuple, _array, _struct, _view, e1;
		err = null;
		_array = new Uint8Array(144);
		_tuple = Syscall(339, (fd >>> 0), _array, 0); e1 = _tuple[2];
		_struct = stat, _view = new DataView(_array.buffer, _array.byteOffset), _struct.Dev = _view.getInt32(0, true), _struct.Mode = _view.getUint16(4, true), _struct.Nlink = _view.getUint16(6, true), _struct.Ino = new Go$Uint64(_view.getUint32(12, true), _view.getUint32(8, true)), _struct.Uid = _view.getUint32(16, true), _struct.Gid = _view.getUint32(20, true), _struct.Rdev = _view.getInt32(24, true), _struct.Pad_cgo_0 = new (go$nativeArray("Uint8"))(_array.buffer, go$min(_array.byteOffset + 28, _array.buffer.byteLength)), _struct.Atimespec.Sec = new Go$Int64(_view.getUint32(36, true), _view.getUint32(32, true)), _struct.Atimespec.Nsec = new Go$Int64(_view.getUint32(44, true), _view.getUint32(40, true)), _struct.Mtimespec.Sec = new Go$Int64(_view.getUint32(52, true), _view.getUint32(48, true)), _struct.Mtimespec.Nsec = new Go$Int64(_view.getUint32(60, true), _view.getUint32(56, true)), _struct.Ctimespec.Sec = new Go$Int64(_view.getUint32(68, true), _view.getUint32(64, true)), _struct.Ctimespec.Nsec = new Go$Int64(_view.getUint32(76, true), _view.getUint32(72, true)), _struct.Birthtimespec.Sec = new Go$Int64(_view.getUint32(84, true), _view.getUint32(80, true)), _struct.Birthtimespec.Nsec = new Go$Int64(_view.getUint32(92, true), _view.getUint32(88, true)), _struct.Size = new Go$Int64(_view.getUint32(100, true), _view.getUint32(96, true)), _struct.Blocks = new Go$Int64(_view.getUint32(108, true), _view.getUint32(104, true)), _struct.Blksize = _view.getInt32(112, true), _struct.Flags = _view.getUint32(116, true), _struct.Gen = _view.getUint32(120, true), _struct.Lspare = _view.getInt32(124, true), _struct.Qspare = new (go$nativeArray("Int64"))(_array.buffer, go$min(_array.byteOffset + 128, _array.buffer.byteLength));
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Fsync = go$pkg.Fsync = function(fd) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(95, (fd >>> 0), 0, 0); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Ftruncate = go$pkg.Ftruncate = function(fd, length) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(201, (fd >>> 0), (length.low >>> 0), 0); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Getdirentries = go$pkg.Getdirentries = function(fd, buf, basep) {
		var n, err, _p0, _tuple, r0, e1;
		n = 0;
		err = null;
		_p0 = 0;
		if (buf.length > 0) {
			_p0 = go$sliceToArray(buf);
		} else {
			_p0 = new Uint8Array(0);
		}
		_tuple = Syscall6(344, (fd >>> 0), _p0, (buf.length >>> 0), basep, 0, 0); r0 = _tuple[0]; e1 = _tuple[2];
		n = (r0 >> 0);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [n, err];
	};
	Lstat = go$pkg.Lstat = function(path, stat) {
		var err, _p0, _tuple, _tuple$1, _array, _struct, _view, e1;
		err = null;
		_p0 = (go$ptrType(Go$Uint8)).nil;
		_tuple = BytePtrFromString(path); _p0 = _tuple[0]; err = _tuple[1];
		if (!(go$interfaceIsEqual(err, null))) {
			return err;
		}
		_array = new Uint8Array(144);
		_tuple$1 = Syscall(340, _p0, _array, 0); e1 = _tuple$1[2];
		_struct = stat, _view = new DataView(_array.buffer, _array.byteOffset), _struct.Dev = _view.getInt32(0, true), _struct.Mode = _view.getUint16(4, true), _struct.Nlink = _view.getUint16(6, true), _struct.Ino = new Go$Uint64(_view.getUint32(12, true), _view.getUint32(8, true)), _struct.Uid = _view.getUint32(16, true), _struct.Gid = _view.getUint32(20, true), _struct.Rdev = _view.getInt32(24, true), _struct.Pad_cgo_0 = new (go$nativeArray("Uint8"))(_array.buffer, go$min(_array.byteOffset + 28, _array.buffer.byteLength)), _struct.Atimespec.Sec = new Go$Int64(_view.getUint32(36, true), _view.getUint32(32, true)), _struct.Atimespec.Nsec = new Go$Int64(_view.getUint32(44, true), _view.getUint32(40, true)), _struct.Mtimespec.Sec = new Go$Int64(_view.getUint32(52, true), _view.getUint32(48, true)), _struct.Mtimespec.Nsec = new Go$Int64(_view.getUint32(60, true), _view.getUint32(56, true)), _struct.Ctimespec.Sec = new Go$Int64(_view.getUint32(68, true), _view.getUint32(64, true)), _struct.Ctimespec.Nsec = new Go$Int64(_view.getUint32(76, true), _view.getUint32(72, true)), _struct.Birthtimespec.Sec = new Go$Int64(_view.getUint32(84, true), _view.getUint32(80, true)), _struct.Birthtimespec.Nsec = new Go$Int64(_view.getUint32(92, true), _view.getUint32(88, true)), _struct.Size = new Go$Int64(_view.getUint32(100, true), _view.getUint32(96, true)), _struct.Blocks = new Go$Int64(_view.getUint32(108, true), _view.getUint32(104, true)), _struct.Blksize = _view.getInt32(112, true), _struct.Flags = _view.getUint32(116, true), _struct.Gen = _view.getUint32(120, true), _struct.Lspare = _view.getInt32(124, true), _struct.Qspare = new (go$nativeArray("Int64"))(_array.buffer, go$min(_array.byteOffset + 128, _array.buffer.byteLength));
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	Open = go$pkg.Open = function(path, mode, perm) {
		var fd, err, _p0, _tuple, _tuple$1, r0, e1;
		fd = 0;
		err = null;
		_p0 = (go$ptrType(Go$Uint8)).nil;
		_tuple = BytePtrFromString(path); _p0 = _tuple[0]; err = _tuple[1];
		if (!(go$interfaceIsEqual(err, null))) {
			return [fd, err];
		}
		_tuple$1 = Syscall(5, _p0, (mode >>> 0), (perm >>> 0)); r0 = _tuple$1[0]; e1 = _tuple$1[2];
		fd = (r0 >> 0);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [fd, err];
	};
	Pread = go$pkg.Pread = function(fd, p, offset) {
		var n, err, _p0, _tuple, r0, e1;
		n = 0;
		err = null;
		_p0 = 0;
		if (p.length > 0) {
			_p0 = go$sliceToArray(p);
		} else {
			_p0 = new Uint8Array(0);
		}
		_tuple = Syscall6(153, (fd >>> 0), _p0, (p.length >>> 0), (offset.low >>> 0), 0, 0); r0 = _tuple[0]; e1 = _tuple[2];
		n = (r0 >> 0);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [n, err];
	};
	Pwrite = go$pkg.Pwrite = function(fd, p, offset) {
		var n, err, _p0, _tuple, r0, e1;
		n = 0;
		err = null;
		_p0 = 0;
		if (p.length > 0) {
			_p0 = go$sliceToArray(p);
		} else {
			_p0 = new Uint8Array(0);
		}
		_tuple = Syscall6(154, (fd >>> 0), _p0, (p.length >>> 0), (offset.low >>> 0), 0, 0); r0 = _tuple[0]; e1 = _tuple[2];
		n = (r0 >> 0);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [n, err];
	};
	read = function(fd, p) {
		var n, err, _p0, _tuple, r0, e1;
		n = 0;
		err = null;
		_p0 = 0;
		if (p.length > 0) {
			_p0 = go$sliceToArray(p);
		} else {
			_p0 = new Uint8Array(0);
		}
		_tuple = Syscall(3, (fd >>> 0), _p0, (p.length >>> 0)); r0 = _tuple[0]; e1 = _tuple[2];
		n = (r0 >> 0);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [n, err];
	};
	Seek = go$pkg.Seek = function(fd, offset, whence) {
		var newoffset, err, _tuple, r0, e1;
		newoffset = new Go$Int64(0, 0);
		err = null;
		_tuple = Syscall(199, (fd >>> 0), (offset.low >>> 0), (whence >>> 0)); r0 = _tuple[0]; e1 = _tuple[2];
		newoffset = new Go$Int64(0, r0.constructor === Number ? r0 : 1);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [newoffset, err];
	};
	write = function(fd, p) {
		var n, err, _p0, _tuple, r0, e1;
		n = 0;
		err = null;
		_p0 = 0;
		if (p.length > 0) {
			_p0 = go$sliceToArray(p);
		} else {
			_p0 = new Uint8Array(0);
		}
		_tuple = Syscall(4, (fd >>> 0), _p0, (p.length >>> 0)); r0 = _tuple[0]; e1 = _tuple[2];
		n = (r0 >> 0);
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [n, err];
	};
	mmap = function(addr, length, prot, flag, fd, pos) {
		var ret, err, _tuple, r0, e1;
		ret = 0;
		err = null;
		_tuple = Syscall6(197, addr, length, (prot >>> 0), (flag >>> 0), (fd >>> 0), (pos.low >>> 0)); r0 = _tuple[0]; e1 = _tuple[2];
		ret = r0;
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return [ret, err];
	};
	munmap = function(addr, length) {
		var err, _tuple, e1;
		err = null;
		_tuple = Syscall(73, addr, length, 0); e1 = _tuple[2];
		if (!((e1 === 0))) {
			err = new Errno(e1);
		}
		return err;
	};
	go$pkg.init = function() {
		(go$ptrType(mmapper)).methods = [["Lock", "", [], [], false, 0], ["Mmap", "", [Go$Int, Go$Int64, Go$Int, Go$Int, Go$Int], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["Munmap", "", [(go$sliceType(Go$Uint8))], [go$error], false, -1], ["Unlock", "", [], [], false, 0]];
		mmapper.init([["Mutex", "", "", sync.Mutex, ""], ["active", "active", "syscall", (go$mapType((go$ptrType(Go$Uint8)), (go$sliceType(Go$Uint8)))), ""], ["mmap", "mmap", "syscall", (go$funcType([Go$Uintptr, Go$Uintptr, Go$Int, Go$Int, Go$Int, Go$Int64], [Go$Uintptr, go$error], false)), ""], ["munmap", "munmap", "syscall", (go$funcType([Go$Uintptr, Go$Uintptr], [go$error], false)), ""]]);
		Errno.methods = [["Error", "", [], [Go$String], false, -1], ["Temporary", "", [], [Go$Bool], false, -1], ["Timeout", "", [], [Go$Bool], false, -1]];
		(go$ptrType(Errno)).methods = [["Error", "", [], [Go$String], false, -1], ["Temporary", "", [], [Go$Bool], false, -1], ["Timeout", "", [], [Go$Bool], false, -1]];
		(go$ptrType(Timespec)).methods = [["Nano", "", [], [Go$Int64], false, -1], ["Unix", "", [], [Go$Int64, Go$Int64], false, -1]];
		Timespec.init([["Sec", "Sec", "", Go$Int64, ""], ["Nsec", "Nsec", "", Go$Int64, ""]]);
		Stat_t.init([["Dev", "Dev", "", Go$Int32, ""], ["Mode", "Mode", "", Go$Uint16, ""], ["Nlink", "Nlink", "", Go$Uint16, ""], ["Ino", "Ino", "", Go$Uint64, ""], ["Uid", "Uid", "", Go$Uint32, ""], ["Gid", "Gid", "", Go$Uint32, ""], ["Rdev", "Rdev", "", Go$Int32, ""], ["Pad_cgo_0", "Pad_cgo_0", "", (go$arrayType(Go$Uint8, 4)), ""], ["Atimespec", "Atimespec", "", Timespec, ""], ["Mtimespec", "Mtimespec", "", Timespec, ""], ["Ctimespec", "Ctimespec", "", Timespec, ""], ["Birthtimespec", "Birthtimespec", "", Timespec, ""], ["Size", "Size", "", Go$Int64, ""], ["Blocks", "Blocks", "", Go$Int64, ""], ["Blksize", "Blksize", "", Go$Int32, ""], ["Flags", "Flags", "", Go$Uint32, ""], ["Gen", "Gen", "", Go$Uint32, ""], ["Lspare", "Lspare", "", Go$Int32, ""], ["Qspare", "Qspare", "", (go$arrayType(Go$Int64, 2)), ""]]);
		Dirent.init([["Ino", "Ino", "", Go$Uint64, ""], ["Seekoff", "Seekoff", "", Go$Uint64, ""], ["Reclen", "Reclen", "", Go$Uint16, ""], ["Namlen", "Namlen", "", Go$Uint16, ""], ["Type", "Type", "", Go$Uint8, ""], ["Name", "Name", "", (go$arrayType(Go$Int8, 1024)), ""], ["Pad_cgo_0", "Pad_cgo_0", "", (go$arrayType(Go$Uint8, 3)), ""]]);
		syscallModule = null;
		envOnce = new sync.Once.Ptr();
		envLock = new sync.RWMutex.Ptr();
		env = false;
		envs = (go$sliceType(Go$String)).nil;
		mapper = new mmapper.Ptr(new sync.Mutex.Ptr(), new Go$Map(), mmap, munmap);
		go$pkg.Stdin = 0;
		go$pkg.Stdout = 1;
		go$pkg.Stderr = 2;
		errors = go$toNativeArray("String", ["", "operation not permitted", "no such file or directory", "no such process", "interrupted system call", "input/output error", "device not configured", "argument list too long", "exec format error", "bad file descriptor", "no child processes", "resource deadlock avoided", "cannot allocate memory", "permission denied", "bad address", "block device required", "resource busy", "file exists", "cross-device link", "operation not supported by device", "not a directory", "is a directory", "invalid argument", "too many open files in system", "too many open files", "inappropriate ioctl for device", "text file busy", "file too large", "no space left on device", "illegal seek", "read-only file system", "too many links", "broken pipe", "numerical argument out of domain", "result too large", "resource temporarily unavailable", "operation now in progress", "operation already in progress", "socket operation on non-socket", "destination address required", "message too long", "protocol wrong type for socket", "protocol not available", "protocol not supported", "socket type not supported", "operation not supported", "protocol family not supported", "address family not supported by protocol family", "address already in use", "can't assign requested address", "network is down", "network is unreachable", "network dropped connection on reset", "software caused connection abort", "connection reset by peer", "no buffer space available", "socket is already connected", "socket is not connected", "can't send after socket shutdown", "too many references: can't splice", "operation timed out", "connection refused", "too many levels of symbolic links", "file name too long", "host is down", "no route to host", "directory not empty", "too many processes", "too many users", "disc quota exceeded", "stale NFS file handle", "too many levels of remote in path", "RPC struct is bad", "RPC version wrong", "RPC prog. not avail", "program version wrong", "bad procedure for program", "no locks available", "function not implemented", "inappropriate file type or format", "authentication error", "need authenticator", "device power is off", "device error", "value too large to be stored in data type", "bad executable (or shared library)", "bad CPU type in executable", "shared library version mismatch", "malformed Mach-o file", "operation canceled", "identifier removed", "no message of desired type", "illegal byte sequence", "attribute not found", "bad message", "EMULTIHOP (Reserved)", "no message available on STREAM", "ENOLINK (Reserved)", "no STREAM resources", "not a STREAM", "protocol error", "STREAM ioctl timeout", "operation not supported on socket", "policy not found", "state not recoverable", "previous owner died"]);
		var process, jsEnv, envkeys, i, key;
		process = go$global.process;
		if (!(process === undefined)) {
			jsEnv = process.env;
			envkeys = go$global.Object.keys(jsEnv);
			envs = (go$sliceType(Go$String)).make(go$parseInt(envkeys.length), 0, function() { return ""; });
			i = 0;
			while (i < go$parseInt(envkeys.length)) {
				key = go$internalize(envkeys[i], Go$String);
				(i < 0 || i >= envs.length) ? go$throwRuntimeError("index out of range") : envs.array[envs.offset + i] = key + "=" + go$internalize(jsEnv[go$externalize(key, Go$String)], Go$String);
				i = i + 1 >> 0;
			}
		}
	}
	return go$pkg;
})();
go$packages["time"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], errors = go$packages["errors"], syscall = go$packages["syscall"], sync = go$packages["sync"], runtime = go$packages["runtime"], ParseError, Time, Month, Weekday, Duration, Location, zone, zoneTrans, data, now, startsWithLowerCase, nextStdChunk, match, lookup, appendUint, atoi, formatNano, quote, isDigit, getnum, cutspace, skip, Parse, parse, parseTimeZone, parseGMT, parseNanoseconds, leadingInt, readFile, open, closefd, preadn, absWeekday, absClock, fmtFrac, fmtInt, absDate, Unix, isLeap, norm, Date, div, FixedZone, byteString, loadZoneData, loadZoneFile, get4, get2, loadZoneZip, initLocal, loadLocation, std0x, longDayNames, shortDayNames, shortMonthNames, longMonthNames, atoiError, errBad, errLeadingInt, months, days, daysBefore, utcLoc, localLoc, localOnce, zoneinfo, badData, zoneDirs;
	ParseError = go$pkg.ParseError = go$newType(0, "Struct", "time.ParseError", "ParseError", "time", function(Layout_, Value_, LayoutElem_, ValueElem_, Message_) {
		this.go$val = this;
		this.Layout = Layout_ !== undefined ? Layout_ : "";
		this.Value = Value_ !== undefined ? Value_ : "";
		this.LayoutElem = LayoutElem_ !== undefined ? LayoutElem_ : "";
		this.ValueElem = ValueElem_ !== undefined ? ValueElem_ : "";
		this.Message = Message_ !== undefined ? Message_ : "";
	});
	Time = go$pkg.Time = go$newType(0, "Struct", "time.Time", "Time", "time", function(sec_, nsec_, loc_) {
		this.go$val = this;
		this.sec = sec_ !== undefined ? sec_ : new Go$Int64(0, 0);
		this.nsec = nsec_ !== undefined ? nsec_ : 0;
		this.loc = loc_ !== undefined ? loc_ : (go$ptrType(Location)).nil;
	});
	Month = go$pkg.Month = go$newType(4, "Int", "time.Month", "Month", "time", null);
	Weekday = go$pkg.Weekday = go$newType(4, "Int", "time.Weekday", "Weekday", "time", null);
	Duration = go$pkg.Duration = go$newType(8, "Int64", "time.Duration", "Duration", "time", null);
	Location = go$pkg.Location = go$newType(0, "Struct", "time.Location", "Location", "time", function(name_, zone_, tx_, cacheStart_, cacheEnd_, cacheZone_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : "";
		this.zone = zone_ !== undefined ? zone_ : (go$sliceType(zone)).nil;
		this.tx = tx_ !== undefined ? tx_ : (go$sliceType(zoneTrans)).nil;
		this.cacheStart = cacheStart_ !== undefined ? cacheStart_ : new Go$Int64(0, 0);
		this.cacheEnd = cacheEnd_ !== undefined ? cacheEnd_ : new Go$Int64(0, 0);
		this.cacheZone = cacheZone_ !== undefined ? cacheZone_ : (go$ptrType(zone)).nil;
	});
	zone = go$pkg.zone = go$newType(0, "Struct", "time.zone", "zone", "time", function(name_, offset_, isDST_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : "";
		this.offset = offset_ !== undefined ? offset_ : 0;
		this.isDST = isDST_ !== undefined ? isDST_ : false;
	});
	zoneTrans = go$pkg.zoneTrans = go$newType(0, "Struct", "time.zoneTrans", "zoneTrans", "time", function(when_, index_, isstd_, isutc_) {
		this.go$val = this;
		this.when = when_ !== undefined ? when_ : new Go$Int64(0, 0);
		this.index = index_ !== undefined ? index_ : 0;
		this.isstd = isstd_ !== undefined ? isstd_ : false;
		this.isutc = isutc_ !== undefined ? isutc_ : false;
	});
	data = go$pkg.data = go$newType(0, "Struct", "time.data", "data", "time", function(p_, error_) {
		this.go$val = this;
		this.p = p_ !== undefined ? p_ : (go$sliceType(Go$Uint8)).nil;
		this.error = error_ !== undefined ? error_ : false;
	});
	now = function() {
		var sec, nsec, msec, _tmp, _tmp$1, x, x$1;
		sec = new Go$Int64(0, 0);
		nsec = 0;
		msec = go$internalize(new (go$global.Date)().getTime(), Go$Int64);
		_tmp = go$div64(msec, new Go$Int64(0, 1000), false); _tmp$1 = (x = ((x$1 = go$div64(msec, new Go$Int64(0, 1000), true), x$1.low + ((x$1.high >> 31) * 4294967296)) >> 0), (((x >>> 16 << 16) * 1000000 >> 0) + (x << 16 >>> 16) * 1000000) >> 0); sec = _tmp; nsec = _tmp$1;
		return [sec, nsec];
	};
	startsWithLowerCase = function(str) {
		var c;
		if (str.length === 0) {
			return false;
		}
		c = str.charCodeAt(0);
		return 97 <= c && c <= 122;
	};
	nextStdChunk = function(layout) {
		var prefix, std, suffix, i, c, _ref, _tmp, _tmp$1, _tmp$2, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, _tmp$8, _tmp$9, _tmp$10, _tmp$11, _tmp$12, _tmp$13, _tmp$14, _tmp$15, _tmp$16, _tmp$17, _tmp$18, _tmp$19, _tmp$20, _tmp$21, _tmp$22, _tmp$23, _tmp$24, _tmp$25, _tmp$26, _tmp$27, _tmp$28, _tmp$29, _tmp$30, _tmp$31, _tmp$32, _tmp$33, _tmp$34, _tmp$35, _tmp$36, _tmp$37, _tmp$38, _tmp$39, _tmp$40, _tmp$41, _tmp$42, _tmp$43, _tmp$44, _tmp$45, _tmp$46, _tmp$47, _tmp$48, _tmp$49, _tmp$50, _tmp$51, _tmp$52, _tmp$53, _tmp$54, _tmp$55, _tmp$56, _tmp$57, _tmp$58, _tmp$59, _tmp$60, _tmp$61, _tmp$62, _tmp$63, _tmp$64, _tmp$65, _tmp$66, _tmp$67, _tmp$68, _tmp$69, _tmp$70, _tmp$71, _tmp$72, _tmp$73, _tmp$74, ch, j, std$1, _tmp$75, _tmp$76, _tmp$77, _tmp$78, _tmp$79, _tmp$80;
		prefix = "";
		std = 0;
		suffix = "";
		i = 0;
		while (i < layout.length) {
			c = (layout.charCodeAt(i) >> 0);
			_ref = c;
			if (_ref === 74) {
				if (layout.length >= (i + 3 >> 0) && layout.substring(i, (i + 3 >> 0)) === "Jan") {
					if (layout.length >= (i + 7 >> 0) && layout.substring(i, (i + 7 >> 0)) === "January") {
						_tmp = layout.substring(0, i); _tmp$1 = 257; _tmp$2 = layout.substring((i + 7 >> 0)); prefix = _tmp; std = _tmp$1; suffix = _tmp$2;
						return [prefix, std, suffix];
					}
					if (!startsWithLowerCase(layout.substring((i + 3 >> 0)))) {
						_tmp$3 = layout.substring(0, i); _tmp$4 = 258; _tmp$5 = layout.substring((i + 3 >> 0)); prefix = _tmp$3; std = _tmp$4; suffix = _tmp$5;
						return [prefix, std, suffix];
					}
				}
			} else if (_ref === 77) {
				if (layout.length >= (i + 3 >> 0)) {
					if (layout.substring(i, (i + 3 >> 0)) === "Mon") {
						if (layout.length >= (i + 6 >> 0) && layout.substring(i, (i + 6 >> 0)) === "Monday") {
							_tmp$6 = layout.substring(0, i); _tmp$7 = 261; _tmp$8 = layout.substring((i + 6 >> 0)); prefix = _tmp$6; std = _tmp$7; suffix = _tmp$8;
							return [prefix, std, suffix];
						}
						if (!startsWithLowerCase(layout.substring((i + 3 >> 0)))) {
							_tmp$9 = layout.substring(0, i); _tmp$10 = 262; _tmp$11 = layout.substring((i + 3 >> 0)); prefix = _tmp$9; std = _tmp$10; suffix = _tmp$11;
							return [prefix, std, suffix];
						}
					}
					if (layout.substring(i, (i + 3 >> 0)) === "MST") {
						_tmp$12 = layout.substring(0, i); _tmp$13 = 21; _tmp$14 = layout.substring((i + 3 >> 0)); prefix = _tmp$12; std = _tmp$13; suffix = _tmp$14;
						return [prefix, std, suffix];
					}
				}
			} else if (_ref === 48) {
				if (layout.length >= (i + 2 >> 0) && 49 <= layout.charCodeAt((i + 1 >> 0)) && layout.charCodeAt((i + 1 >> 0)) <= 54) {
					_tmp$15 = layout.substring(0, i); _tmp$16 = std0x[(layout.charCodeAt((i + 1 >> 0)) - 49 << 24 >>> 24)]; _tmp$17 = layout.substring((i + 2 >> 0)); prefix = _tmp$15; std = _tmp$16; suffix = _tmp$17;
					return [prefix, std, suffix];
				}
			} else if (_ref === 49) {
				if (layout.length >= (i + 2 >> 0) && (layout.charCodeAt((i + 1 >> 0)) === 53)) {
					_tmp$18 = layout.substring(0, i); _tmp$19 = 522; _tmp$20 = layout.substring((i + 2 >> 0)); prefix = _tmp$18; std = _tmp$19; suffix = _tmp$20;
					return [prefix, std, suffix];
				}
				_tmp$21 = layout.substring(0, i); _tmp$22 = 259; _tmp$23 = layout.substring((i + 1 >> 0)); prefix = _tmp$21; std = _tmp$22; suffix = _tmp$23;
				return [prefix, std, suffix];
			} else if (_ref === 50) {
				if (layout.length >= (i + 4 >> 0) && layout.substring(i, (i + 4 >> 0)) === "2006") {
					_tmp$24 = layout.substring(0, i); _tmp$25 = 273; _tmp$26 = layout.substring((i + 4 >> 0)); prefix = _tmp$24; std = _tmp$25; suffix = _tmp$26;
					return [prefix, std, suffix];
				}
				_tmp$27 = layout.substring(0, i); _tmp$28 = 263; _tmp$29 = layout.substring((i + 1 >> 0)); prefix = _tmp$27; std = _tmp$28; suffix = _tmp$29;
				return [prefix, std, suffix];
			} else if (_ref === 95) {
				if (layout.length >= (i + 2 >> 0) && (layout.charCodeAt((i + 1 >> 0)) === 50)) {
					_tmp$30 = layout.substring(0, i); _tmp$31 = 264; _tmp$32 = layout.substring((i + 2 >> 0)); prefix = _tmp$30; std = _tmp$31; suffix = _tmp$32;
					return [prefix, std, suffix];
				}
			} else if (_ref === 51) {
				_tmp$33 = layout.substring(0, i); _tmp$34 = 523; _tmp$35 = layout.substring((i + 1 >> 0)); prefix = _tmp$33; std = _tmp$34; suffix = _tmp$35;
				return [prefix, std, suffix];
			} else if (_ref === 52) {
				_tmp$36 = layout.substring(0, i); _tmp$37 = 525; _tmp$38 = layout.substring((i + 1 >> 0)); prefix = _tmp$36; std = _tmp$37; suffix = _tmp$38;
				return [prefix, std, suffix];
			} else if (_ref === 53) {
				_tmp$39 = layout.substring(0, i); _tmp$40 = 527; _tmp$41 = layout.substring((i + 1 >> 0)); prefix = _tmp$39; std = _tmp$40; suffix = _tmp$41;
				return [prefix, std, suffix];
			} else if (_ref === 80) {
				if (layout.length >= (i + 2 >> 0) && (layout.charCodeAt((i + 1 >> 0)) === 77)) {
					_tmp$42 = layout.substring(0, i); _tmp$43 = 531; _tmp$44 = layout.substring((i + 2 >> 0)); prefix = _tmp$42; std = _tmp$43; suffix = _tmp$44;
					return [prefix, std, suffix];
				}
			} else if (_ref === 112) {
				if (layout.length >= (i + 2 >> 0) && (layout.charCodeAt((i + 1 >> 0)) === 109)) {
					_tmp$45 = layout.substring(0, i); _tmp$46 = 532; _tmp$47 = layout.substring((i + 2 >> 0)); prefix = _tmp$45; std = _tmp$46; suffix = _tmp$47;
					return [prefix, std, suffix];
				}
			} else if (_ref === 45) {
				if (layout.length >= (i + 7 >> 0) && layout.substring(i, (i + 7 >> 0)) === "-070000") {
					_tmp$48 = layout.substring(0, i); _tmp$49 = 27; _tmp$50 = layout.substring((i + 7 >> 0)); prefix = _tmp$48; std = _tmp$49; suffix = _tmp$50;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 9 >> 0) && layout.substring(i, (i + 9 >> 0)) === "-07:00:00") {
					_tmp$51 = layout.substring(0, i); _tmp$52 = 30; _tmp$53 = layout.substring((i + 9 >> 0)); prefix = _tmp$51; std = _tmp$52; suffix = _tmp$53;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 5 >> 0) && layout.substring(i, (i + 5 >> 0)) === "-0700") {
					_tmp$54 = layout.substring(0, i); _tmp$55 = 26; _tmp$56 = layout.substring((i + 5 >> 0)); prefix = _tmp$54; std = _tmp$55; suffix = _tmp$56;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 6 >> 0) && layout.substring(i, (i + 6 >> 0)) === "-07:00") {
					_tmp$57 = layout.substring(0, i); _tmp$58 = 29; _tmp$59 = layout.substring((i + 6 >> 0)); prefix = _tmp$57; std = _tmp$58; suffix = _tmp$59;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 3 >> 0) && layout.substring(i, (i + 3 >> 0)) === "-07") {
					_tmp$60 = layout.substring(0, i); _tmp$61 = 28; _tmp$62 = layout.substring((i + 3 >> 0)); prefix = _tmp$60; std = _tmp$61; suffix = _tmp$62;
					return [prefix, std, suffix];
				}
			} else if (_ref === 90) {
				if (layout.length >= (i + 7 >> 0) && layout.substring(i, (i + 7 >> 0)) === "Z070000") {
					_tmp$63 = layout.substring(0, i); _tmp$64 = 23; _tmp$65 = layout.substring((i + 7 >> 0)); prefix = _tmp$63; std = _tmp$64; suffix = _tmp$65;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 9 >> 0) && layout.substring(i, (i + 9 >> 0)) === "Z07:00:00") {
					_tmp$66 = layout.substring(0, i); _tmp$67 = 25; _tmp$68 = layout.substring((i + 9 >> 0)); prefix = _tmp$66; std = _tmp$67; suffix = _tmp$68;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 5 >> 0) && layout.substring(i, (i + 5 >> 0)) === "Z0700") {
					_tmp$69 = layout.substring(0, i); _tmp$70 = 22; _tmp$71 = layout.substring((i + 5 >> 0)); prefix = _tmp$69; std = _tmp$70; suffix = _tmp$71;
					return [prefix, std, suffix];
				}
				if (layout.length >= (i + 6 >> 0) && layout.substring(i, (i + 6 >> 0)) === "Z07:00") {
					_tmp$72 = layout.substring(0, i); _tmp$73 = 24; _tmp$74 = layout.substring((i + 6 >> 0)); prefix = _tmp$72; std = _tmp$73; suffix = _tmp$74;
					return [prefix, std, suffix];
				}
			} else if (_ref === 46) {
				if ((i + 1 >> 0) < layout.length && ((layout.charCodeAt((i + 1 >> 0)) === 48) || (layout.charCodeAt((i + 1 >> 0)) === 57))) {
					ch = layout.charCodeAt((i + 1 >> 0));
					j = i + 1 >> 0;
					while (j < layout.length && (layout.charCodeAt(j) === ch)) {
						j = j + 1 >> 0;
					}
					if (!isDigit(layout, j)) {
						std$1 = 31;
						if (layout.charCodeAt((i + 1 >> 0)) === 57) {
							std$1 = 32;
						}
						std$1 = std$1 | ((((j - ((i + 1 >> 0)) >> 0)) << 16 >> 0));
						_tmp$75 = layout.substring(0, i); _tmp$76 = std$1; _tmp$77 = layout.substring(j); prefix = _tmp$75; std = _tmp$76; suffix = _tmp$77;
						return [prefix, std, suffix];
					}
				}
			}
			i = i + 1 >> 0;
		}
		_tmp$78 = layout; _tmp$79 = 0; _tmp$80 = ""; prefix = _tmp$78; std = _tmp$79; suffix = _tmp$80;
		return [prefix, std, suffix];
	};
	match = function(s1, s2) {
		var i, c1, c2;
		i = 0;
		while (i < s1.length) {
			c1 = s1.charCodeAt(i);
			c2 = s2.charCodeAt(i);
			if (!((c1 === c2))) {
				c1 = (c1 | 32) >>> 0;
				c2 = (c2 | 32) >>> 0;
				if (!((c1 === c2)) || c1 < 97 || c1 > 122) {
					return false;
				}
			}
			i = i + 1 >> 0;
		}
		return true;
	};
	lookup = function(tab, val) {
		var _ref, _i, v, i;
		_ref = tab;
		_i = 0;
		while (_i < _ref.length) {
			v = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if (val.length >= v.length && match(val.substring(0, v.length), v)) {
				return [i, val.substring(v.length), null];
			}
			_i++;
		}
		return [-1, val, errBad];
	};
	appendUint = function(b, x, pad) {
		var _q, _r, buf, n, _r$1, _q$1;
		if (x < 10) {
			if (!((pad === 0))) {
				b = go$append(b, pad);
			}
			return go$append(b, ((48 + x >>> 0) << 24 >>> 24));
		}
		if (x < 100) {
			b = go$append(b, ((48 + (_q = x / 10, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >>> 0 : go$throwRuntimeError("integer divide by zero")) >>> 0) << 24 >>> 24));
			b = go$append(b, ((48 + (_r = x % 10, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) >>> 0) << 24 >>> 24));
			return b;
		}
		buf = go$makeNativeArray("Uint8", 32, function() { return 0; });
		n = 32;
		if (x === 0) {
			return go$append(b, 48);
		}
		while (x >= 10) {
			n = n - 1 >> 0;
			buf[n] = (((_r$1 = x % 10, _r$1 === _r$1 ? _r$1 : go$throwRuntimeError("integer divide by zero")) + 48 >>> 0) << 24 >>> 24);
			x = (_q$1 = x / 10, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >>> 0 : go$throwRuntimeError("integer divide by zero"));
		}
		n = n - 1 >> 0;
		buf[n] = ((x + 48 >>> 0) << 24 >>> 24);
		return go$appendSlice(b, go$subslice(new (go$sliceType(Go$Uint8))(buf), n));
	};
	atoi = function(s) {
		var x, err, neg, _tuple, q, rem, _tmp, _tmp$1, _tmp$2, _tmp$3;
		x = 0;
		err = null;
		neg = false;
		if (!(s === "") && ((s.charCodeAt(0) === 45) || (s.charCodeAt(0) === 43))) {
			neg = s.charCodeAt(0) === 45;
			s = s.substring(1);
		}
		_tuple = leadingInt(s); q = _tuple[0]; rem = _tuple[1]; err = _tuple[2];
		x = ((q.low + ((q.high >> 31) * 4294967296)) >> 0);
		if (!(go$interfaceIsEqual(err, null)) || !(rem === "")) {
			_tmp = 0; _tmp$1 = atoiError; x = _tmp; err = _tmp$1;
			return [x, err];
		}
		if (neg) {
			x = -x;
		}
		_tmp$2 = x; _tmp$3 = null; x = _tmp$2; err = _tmp$3;
		return [x, err];
	};
	formatNano = function(b, nanosec, n, trim) {
		var u, buf, start, _r, _q;
		u = nanosec;
		buf = go$makeNativeArray("Uint8", 9, function() { return 0; });
		start = 9;
		while (start > 0) {
			start = start - 1 >> 0;
			buf[start] = (((_r = u % 10, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) + 48 >>> 0) << 24 >>> 24);
			u = (_q = u / 10, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >>> 0 : go$throwRuntimeError("integer divide by zero"));
		}
		if (n > 9) {
			n = 9;
		}
		if (trim) {
			while (n > 0 && (buf[(n - 1 >> 0)] === 48)) {
				n = n - 1 >> 0;
			}
			if (n === 0) {
				return b;
			}
		}
		b = go$append(b, 46);
		return go$appendSlice(b, go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, n));
	};
	Time.Ptr.prototype.String = function() {
		var _struct, t;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return t.Format("2006-01-02 15:04:05.999999999 -0700 MST");
	};
	Time.prototype.String = function() { return this.go$val.String(); };
	Time.Ptr.prototype.Format = function(layout) {
		var _struct, t, _tuple, name, offset, abs, year, month, day, hour, min, sec, b, buf, max, _tuple$1, prefix, std, suffix, _tuple$2, _tuple$3, _ref, y, _r, y$1, m, s, _r$1, hr, _r$2, hr$1, _q, zone$1, absoffset, _q$1, _r$3, _r$4, _q$2, zone$2, _q$3, _r$5;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.locabs(); name = _tuple[0]; offset = _tuple[1]; abs = _tuple[2];
		year = -1;
		month = 0;
		day = 0;
		hour = -1;
		min = 0;
		sec = 0;
		b = (go$sliceType(Go$Uint8)).nil;
		buf = go$makeNativeArray("Uint8", 64, function() { return 0; });
		max = layout.length + 10 >> 0;
		if (max <= 64) {
			b = go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, 0);
		} else {
			b = (go$sliceType(Go$Uint8)).make(0, max, function() { return 0; });
		}
		while (!(layout === "")) {
			_tuple$1 = nextStdChunk(layout); prefix = _tuple$1[0]; std = _tuple$1[1]; suffix = _tuple$1[2];
			if (!(prefix === "")) {
				b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes(prefix)));
			}
			if (std === 0) {
				break;
			}
			layout = suffix;
			if (year < 0 && !(((std & 256) === 0))) {
				_tuple$2 = absDate(abs, true); year = _tuple$2[0]; month = _tuple$2[1]; day = _tuple$2[2];
			}
			if (hour < 0 && !(((std & 512) === 0))) {
				_tuple$3 = absClock(abs); hour = _tuple$3[0]; min = _tuple$3[1]; sec = _tuple$3[2];
			}
			_ref = std & 65535;
			switch (0) { default: if (_ref === 274) {
				y = year;
				if (y < 0) {
					y = -y;
				}
				b = appendUint(b, ((_r = y % 100, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) >>> 0), 48);
			} else if (_ref === 273) {
				y$1 = year;
				if (year <= -1000) {
					b = go$append(b, 45);
					y$1 = -y$1;
				} else if (year <= -100) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("-0")));
					y$1 = -y$1;
				} else if (year <= -10) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("-00")));
					y$1 = -y$1;
				} else if (year < 0) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("-000")));
					y$1 = -y$1;
				} else if (year < 10) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("000")));
				} else if (year < 100) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("00")));
				} else if (year < 1000) {
					b = go$append(b, 48);
				}
				b = appendUint(b, (y$1 >>> 0), 0);
			} else if (_ref === 258) {
				b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes((new Month(month)).String().substring(0, 3))));
			} else if (_ref === 257) {
				m = (new Month(month)).String();
				b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes(m)));
			} else if (_ref === 259) {
				b = appendUint(b, (month >>> 0), 0);
			} else if (_ref === 260) {
				b = appendUint(b, (month >>> 0), 48);
			} else if (_ref === 262) {
				b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes((new Weekday(absWeekday(abs))).String().substring(0, 3))));
			} else if (_ref === 261) {
				s = (new Weekday(absWeekday(abs))).String();
				b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes(s)));
			} else if (_ref === 263) {
				b = appendUint(b, (day >>> 0), 0);
			} else if (_ref === 264) {
				b = appendUint(b, (day >>> 0), 32);
			} else if (_ref === 265) {
				b = appendUint(b, (day >>> 0), 48);
			} else if (_ref === 522) {
				b = appendUint(b, (hour >>> 0), 48);
			} else if (_ref === 523) {
				hr = (_r$1 = hour % 12, _r$1 === _r$1 ? _r$1 : go$throwRuntimeError("integer divide by zero"));
				if (hr === 0) {
					hr = 12;
				}
				b = appendUint(b, (hr >>> 0), 0);
			} else if (_ref === 524) {
				hr$1 = (_r$2 = hour % 12, _r$2 === _r$2 ? _r$2 : go$throwRuntimeError("integer divide by zero"));
				if (hr$1 === 0) {
					hr$1 = 12;
				}
				b = appendUint(b, (hr$1 >>> 0), 48);
			} else if (_ref === 525) {
				b = appendUint(b, (min >>> 0), 0);
			} else if (_ref === 526) {
				b = appendUint(b, (min >>> 0), 48);
			} else if (_ref === 527) {
				b = appendUint(b, (sec >>> 0), 0);
			} else if (_ref === 528) {
				b = appendUint(b, (sec >>> 0), 48);
			} else if (_ref === 531) {
				if (hour >= 12) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("PM")));
				} else {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("AM")));
				}
			} else if (_ref === 532) {
				if (hour >= 12) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("pm")));
				} else {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes("am")));
				}
			} else if (_ref === 22 || _ref === 24 || _ref === 23 || _ref === 25 || _ref === 26 || _ref === 29 || _ref === 27 || _ref === 30) {
				if ((offset === 0) && ((std === 22) || (std === 24) || (std === 23) || (std === 25))) {
					b = go$append(b, 90);
					break;
				}
				zone$1 = (_q = offset / 60, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
				absoffset = offset;
				if (zone$1 < 0) {
					b = go$append(b, 45);
					zone$1 = -zone$1;
					absoffset = -absoffset;
				} else {
					b = go$append(b, 43);
				}
				b = appendUint(b, ((_q$1 = zone$1 / 60, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >> 0 : go$throwRuntimeError("integer divide by zero")) >>> 0), 48);
				if ((std === 24) || (std === 29)) {
					b = go$append(b, 58);
				}
				b = appendUint(b, ((_r$3 = zone$1 % 60, _r$3 === _r$3 ? _r$3 : go$throwRuntimeError("integer divide by zero")) >>> 0), 48);
				if ((std === 23) || (std === 27) || (std === 30) || (std === 25)) {
					if ((std === 30) || (std === 25)) {
						b = go$append(b, 58);
					}
					b = appendUint(b, ((_r$4 = absoffset % 60, _r$4 === _r$4 ? _r$4 : go$throwRuntimeError("integer divide by zero")) >>> 0), 48);
				}
			} else if (_ref === 21) {
				if (!(name === "")) {
					b = go$appendSlice(b, new (go$sliceType(Go$Uint8))(go$stringToBytes(name)));
					break;
				}
				zone$2 = (_q$2 = offset / 60, (_q$2 === _q$2 && _q$2 !== 1/0 && _q$2 !== -1/0) ? _q$2 >> 0 : go$throwRuntimeError("integer divide by zero"));
				if (zone$2 < 0) {
					b = go$append(b, 45);
					zone$2 = -zone$2;
				} else {
					b = go$append(b, 43);
				}
				b = appendUint(b, ((_q$3 = zone$2 / 60, (_q$3 === _q$3 && _q$3 !== 1/0 && _q$3 !== -1/0) ? _q$3 >> 0 : go$throwRuntimeError("integer divide by zero")) >>> 0), 48);
				b = appendUint(b, ((_r$5 = zone$2 % 60, _r$5 === _r$5 ? _r$5 : go$throwRuntimeError("integer divide by zero")) >>> 0), 48);
			} else if (_ref === 31 || _ref === 32) {
				b = formatNano(b, (t.Nanosecond() >>> 0), std >> 16 >> 0, (std & 65535) === 32);
			} }
		}
		return go$bytesToString(b);
	};
	Time.prototype.Format = function(layout) { return this.go$val.Format(layout); };
	quote = function(s) {
		return "\"" + s + "\"";
	};
	ParseError.Ptr.prototype.Error = function() {
		var e;
		e = this;
		if (e.Message === "") {
			return "parsing time " + quote(e.Value) + " as " + quote(e.Layout) + ": cannot parse " + quote(e.ValueElem) + " as " + quote(e.LayoutElem);
		}
		return "parsing time " + quote(e.Value) + e.Message;
	};
	ParseError.prototype.Error = function() { return this.go$val.Error(); };
	isDigit = function(s, i) {
		var c;
		if (s.length <= i) {
			return false;
		}
		c = s.charCodeAt(i);
		return 48 <= c && c <= 57;
	};
	getnum = function(s, fixed) {
		var x;
		if (!isDigit(s, 0)) {
			return [0, s, errBad];
		}
		if (!isDigit(s, 1)) {
			if (fixed) {
				return [0, s, errBad];
			}
			return [((s.charCodeAt(0) - 48 << 24 >>> 24) >> 0), s.substring(1), null];
		}
		return [(x = ((s.charCodeAt(0) - 48 << 24 >>> 24) >> 0), (((x >>> 16 << 16) * 10 >> 0) + (x << 16 >>> 16) * 10) >> 0) + ((s.charCodeAt(1) - 48 << 24 >>> 24) >> 0) >> 0, s.substring(2), null];
	};
	cutspace = function(s) {
		while (s.length > 0 && (s.charCodeAt(0) === 32)) {
			s = s.substring(1);
		}
		return s;
	};
	skip = function(value, prefix) {
		while (prefix.length > 0) {
			if (prefix.charCodeAt(0) === 32) {
				if (value.length > 0 && !((value.charCodeAt(0) === 32))) {
					return [value, errBad];
				}
				prefix = cutspace(prefix);
				value = cutspace(value);
				continue;
			}
			if ((value.length === 0) || !((value.charCodeAt(0) === prefix.charCodeAt(0)))) {
				return [value, errBad];
			}
			prefix = prefix.substring(1);
			value = value.substring(1);
		}
		return [value, null];
	};
	Parse = go$pkg.Parse = function(layout, value) {
		return parse(layout, value, go$pkg.UTC, go$pkg.Local);
	};
	parse = function(layout, value, defaultLocation, local) {
		var _tmp, _tmp$1, alayout, avalue, rangeErrString, amSet, pmSet, year, month, day, hour, min, sec, nsec, z, zoneOffset, zoneName, err, _tuple, prefix, std, suffix, stdstr, _tuple$1, p, _ref, _tmp$2, _tmp$3, _tuple$2, _tmp$4, _tmp$5, _tuple$3, _tuple$4, _tuple$5, _tuple$6, _tuple$7, _tuple$8, _tuple$9, _tuple$10, _tuple$11, _tuple$12, _tuple$13, _tuple$14, n, _tuple$15, _tmp$6, _tmp$7, _ref$1, _tmp$8, _tmp$9, _ref$2, _tmp$10, _tmp$11, _tmp$12, _tmp$13, sign, hour$1, min$1, seconds, _tmp$14, _tmp$15, _tmp$16, _tmp$17, _tmp$18, _tmp$19, _tmp$20, _tmp$21, _tmp$22, _tmp$23, _tmp$24, _tmp$25, _tmp$26, _tmp$27, _tmp$28, _tmp$29, _tmp$30, _tmp$31, _tmp$32, _tmp$33, _tmp$34, _tmp$35, _tmp$36, _tmp$37, _tmp$38, _tmp$39, _tmp$40, _tmp$41, hr, mm, ss, _tuple$16, _tuple$17, _tuple$18, x, _ref$3, _tuple$19, n$1, ok, _tmp$42, _tmp$43, ndigit, _tuple$20, i, _tuple$21, _struct, _struct$1, t, x$1, x$2, _tuple$22, x$3, name, offset, _struct$2, _struct$3, _struct$4, t$1, _tuple$23, x$4, offset$1, ok$1, x$5, x$6, _struct$5, _tuple$24, _struct$6, _struct$7;
		_tmp = layout; _tmp$1 = value; alayout = _tmp; avalue = _tmp$1;
		rangeErrString = "";
		amSet = false;
		pmSet = false;
		year = 0;
		month = 1;
		day = 1;
		hour = 0;
		min = 0;
		sec = 0;
		nsec = 0;
		z = (go$ptrType(Location)).nil;
		zoneOffset = -1;
		zoneName = "";
		while (true) {
			err = null;
			_tuple = nextStdChunk(layout); prefix = _tuple[0]; std = _tuple[1]; suffix = _tuple[2];
			stdstr = layout.substring(prefix.length, (layout.length - suffix.length >> 0));
			_tuple$1 = skip(value, prefix); value = _tuple$1[0]; err = _tuple$1[1];
			if (!(go$interfaceIsEqual(err, null))) {
				return [new Time.Ptr(new Go$Int64(0, 0), 0, (go$ptrType(Location)).nil), new ParseError.Ptr(alayout, avalue, prefix, value, "")];
			}
			if (std === 0) {
				if (!((value.length === 0))) {
					return [new Time.Ptr(new Go$Int64(0, 0), 0, (go$ptrType(Location)).nil), new ParseError.Ptr(alayout, avalue, "", value, ": extra text: " + value)];
				}
				break;
			}
			layout = suffix;
			p = "";
			_ref = std & 65535;
			switch (0) { default: if (_ref === 274) {
				if (value.length < 2) {
					err = errBad;
					break;
				}
				_tmp$2 = value.substring(0, 2); _tmp$3 = value.substring(2); p = _tmp$2; value = _tmp$3;
				_tuple$2 = atoi(p); year = _tuple$2[0]; err = _tuple$2[1];
				if (year >= 69) {
					year = year + 1900 >> 0;
				} else {
					year = year + 2000 >> 0;
				}
			} else if (_ref === 273) {
				if (value.length < 4 || !isDigit(value, 0)) {
					err = errBad;
					break;
				}
				_tmp$4 = value.substring(0, 4); _tmp$5 = value.substring(4); p = _tmp$4; value = _tmp$5;
				_tuple$3 = atoi(p); year = _tuple$3[0]; err = _tuple$3[1];
			} else if (_ref === 258) {
				_tuple$4 = lookup(shortMonthNames, value); month = _tuple$4[0]; value = _tuple$4[1]; err = _tuple$4[2];
			} else if (_ref === 257) {
				_tuple$5 = lookup(longMonthNames, value); month = _tuple$5[0]; value = _tuple$5[1]; err = _tuple$5[2];
			} else if (_ref === 259 || _ref === 260) {
				_tuple$6 = getnum(value, std === 260); month = _tuple$6[0]; value = _tuple$6[1]; err = _tuple$6[2];
				if (month <= 0 || 12 < month) {
					rangeErrString = "month";
				}
			} else if (_ref === 262) {
				_tuple$7 = lookup(shortDayNames, value); value = _tuple$7[1]; err = _tuple$7[2];
			} else if (_ref === 261) {
				_tuple$8 = lookup(longDayNames, value); value = _tuple$8[1]; err = _tuple$8[2];
			} else if (_ref === 263 || _ref === 264 || _ref === 265) {
				if ((std === 264) && value.length > 0 && (value.charCodeAt(0) === 32)) {
					value = value.substring(1);
				}
				_tuple$9 = getnum(value, std === 265); day = _tuple$9[0]; value = _tuple$9[1]; err = _tuple$9[2];
				if (day < 0 || 31 < day) {
					rangeErrString = "day";
				}
			} else if (_ref === 522) {
				_tuple$10 = getnum(value, false); hour = _tuple$10[0]; value = _tuple$10[1]; err = _tuple$10[2];
				if (hour < 0 || 24 <= hour) {
					rangeErrString = "hour";
				}
			} else if (_ref === 523 || _ref === 524) {
				_tuple$11 = getnum(value, std === 524); hour = _tuple$11[0]; value = _tuple$11[1]; err = _tuple$11[2];
				if (hour < 0 || 12 < hour) {
					rangeErrString = "hour";
				}
			} else if (_ref === 525 || _ref === 526) {
				_tuple$12 = getnum(value, std === 526); min = _tuple$12[0]; value = _tuple$12[1]; err = _tuple$12[2];
				if (min < 0 || 60 <= min) {
					rangeErrString = "minute";
				}
			} else if (_ref === 527 || _ref === 528) {
				_tuple$13 = getnum(value, std === 528); sec = _tuple$13[0]; value = _tuple$13[1]; err = _tuple$13[2];
				if (sec < 0 || 60 <= sec) {
					rangeErrString = "second";
				}
				if (value.length >= 2 && (value.charCodeAt(0) === 46) && isDigit(value, 1)) {
					_tuple$14 = nextStdChunk(layout); std = _tuple$14[1];
					std = std & 65535;
					if ((std === 31) || (std === 32)) {
						break;
					}
					n = 2;
					while (n < value.length && isDigit(value, n)) {
						n = n + 1 >> 0;
					}
					_tuple$15 = parseNanoseconds(value, n); nsec = _tuple$15[0]; rangeErrString = _tuple$15[1]; err = _tuple$15[2];
					value = value.substring(n);
				}
			} else if (_ref === 531) {
				if (value.length < 2) {
					err = errBad;
					break;
				}
				_tmp$6 = value.substring(0, 2); _tmp$7 = value.substring(2); p = _tmp$6; value = _tmp$7;
				_ref$1 = p;
				if (_ref$1 === "PM") {
					pmSet = true;
				} else if (_ref$1 === "AM") {
					amSet = true;
				} else {
					err = errBad;
				}
			} else if (_ref === 532) {
				if (value.length < 2) {
					err = errBad;
					break;
				}
				_tmp$8 = value.substring(0, 2); _tmp$9 = value.substring(2); p = _tmp$8; value = _tmp$9;
				_ref$2 = p;
				if (_ref$2 === "pm") {
					pmSet = true;
				} else if (_ref$2 === "am") {
					amSet = true;
				} else {
					err = errBad;
				}
			} else if (_ref === 22 || _ref === 24 || _ref === 23 || _ref === 25 || _ref === 26 || _ref === 28 || _ref === 29 || _ref === 27 || _ref === 30) {
				if (((std === 22) || (std === 24)) && value.length >= 1 && (value.charCodeAt(0) === 90)) {
					value = value.substring(1);
					z = go$pkg.UTC;
					break;
				}
				_tmp$10 = ""; _tmp$11 = ""; _tmp$12 = ""; _tmp$13 = ""; sign = _tmp$10; hour$1 = _tmp$11; min$1 = _tmp$12; seconds = _tmp$13;
				if ((std === 24) || (std === 29)) {
					if (value.length < 6) {
						err = errBad;
						break;
					}
					if (!((value.charCodeAt(3) === 58))) {
						err = errBad;
						break;
					}
					_tmp$14 = value.substring(0, 1); _tmp$15 = value.substring(1, 3); _tmp$16 = value.substring(4, 6); _tmp$17 = "00"; _tmp$18 = value.substring(6); sign = _tmp$14; hour$1 = _tmp$15; min$1 = _tmp$16; seconds = _tmp$17; value = _tmp$18;
				} else if (std === 28) {
					if (value.length < 3) {
						err = errBad;
						break;
					}
					_tmp$19 = value.substring(0, 1); _tmp$20 = value.substring(1, 3); _tmp$21 = "00"; _tmp$22 = "00"; _tmp$23 = value.substring(3); sign = _tmp$19; hour$1 = _tmp$20; min$1 = _tmp$21; seconds = _tmp$22; value = _tmp$23;
				} else if ((std === 25) || (std === 30)) {
					if (value.length < 9) {
						err = errBad;
						break;
					}
					if (!((value.charCodeAt(3) === 58)) || !((value.charCodeAt(6) === 58))) {
						err = errBad;
						break;
					}
					_tmp$24 = value.substring(0, 1); _tmp$25 = value.substring(1, 3); _tmp$26 = value.substring(4, 6); _tmp$27 = value.substring(7, 9); _tmp$28 = value.substring(9); sign = _tmp$24; hour$1 = _tmp$25; min$1 = _tmp$26; seconds = _tmp$27; value = _tmp$28;
				} else if ((std === 23) || (std === 27)) {
					if (value.length < 7) {
						err = errBad;
						break;
					}
					_tmp$29 = value.substring(0, 1); _tmp$30 = value.substring(1, 3); _tmp$31 = value.substring(3, 5); _tmp$32 = value.substring(5, 7); _tmp$33 = value.substring(7); sign = _tmp$29; hour$1 = _tmp$30; min$1 = _tmp$31; seconds = _tmp$32; value = _tmp$33;
				} else {
					if (value.length < 5) {
						err = errBad;
						break;
					}
					_tmp$34 = value.substring(0, 1); _tmp$35 = value.substring(1, 3); _tmp$36 = value.substring(3, 5); _tmp$37 = "00"; _tmp$38 = value.substring(5); sign = _tmp$34; hour$1 = _tmp$35; min$1 = _tmp$36; seconds = _tmp$37; value = _tmp$38;
				}
				_tmp$39 = 0; _tmp$40 = 0; _tmp$41 = 0; hr = _tmp$39; mm = _tmp$40; ss = _tmp$41;
				_tuple$16 = atoi(hour$1); hr = _tuple$16[0]; err = _tuple$16[1];
				if (go$interfaceIsEqual(err, null)) {
					_tuple$17 = atoi(min$1); mm = _tuple$17[0]; err = _tuple$17[1];
				}
				if (go$interfaceIsEqual(err, null)) {
					_tuple$18 = atoi(seconds); ss = _tuple$18[0]; err = _tuple$18[1];
				}
				zoneOffset = (x = (((((hr >>> 16 << 16) * 60 >> 0) + (hr << 16 >>> 16) * 60) >> 0) + mm >> 0), (((x >>> 16 << 16) * 60 >> 0) + (x << 16 >>> 16) * 60) >> 0) + ss >> 0;
				_ref$3 = sign.charCodeAt(0);
				if (_ref$3 === 43) {
				} else if (_ref$3 === 45) {
					zoneOffset = -zoneOffset;
				} else {
					err = errBad;
				}
			} else if (_ref === 21) {
				if (value.length >= 3 && value.substring(0, 3) === "UTC") {
					z = go$pkg.UTC;
					value = value.substring(3);
					break;
				}
				_tuple$19 = parseTimeZone(value); n$1 = _tuple$19[0]; ok = _tuple$19[1];
				if (!ok) {
					err = errBad;
					break;
				}
				_tmp$42 = value.substring(0, n$1); _tmp$43 = value.substring(n$1); zoneName = _tmp$42; value = _tmp$43;
			} else if (_ref === 31) {
				ndigit = 1 + ((std >> 16 >> 0)) >> 0;
				if (value.length < ndigit) {
					err = errBad;
					break;
				}
				_tuple$20 = parseNanoseconds(value, ndigit); nsec = _tuple$20[0]; rangeErrString = _tuple$20[1]; err = _tuple$20[2];
				value = value.substring(ndigit);
			} else if (_ref === 32) {
				if (value.length < 2 || !((value.charCodeAt(0) === 46)) || value.charCodeAt(1) < 48 || 57 < value.charCodeAt(1)) {
					break;
				}
				i = 0;
				while (i < 9 && (i + 1 >> 0) < value.length && 48 <= value.charCodeAt((i + 1 >> 0)) && value.charCodeAt((i + 1 >> 0)) <= 57) {
					i = i + 1 >> 0;
				}
				_tuple$21 = parseNanoseconds(value, 1 + i >> 0); nsec = _tuple$21[0]; rangeErrString = _tuple$21[1]; err = _tuple$21[2];
				value = value.substring((1 + i >> 0));
			} }
			if (!(rangeErrString === "")) {
				return [new Time.Ptr(new Go$Int64(0, 0), 0, (go$ptrType(Location)).nil), new ParseError.Ptr(alayout, avalue, stdstr, value, ": " + rangeErrString + " out of range")];
			}
			if (!(go$interfaceIsEqual(err, null))) {
				return [new Time.Ptr(new Go$Int64(0, 0), 0, (go$ptrType(Location)).nil), new ParseError.Ptr(alayout, avalue, stdstr, value, "")];
			}
		}
		if (pmSet && hour < 12) {
			hour = hour + 12 >> 0;
		} else if (amSet && (hour === 12)) {
			hour = 0;
		}
		if (!(z === (go$ptrType(Location)).nil)) {
			return [(_struct = Date(year, (month >> 0), day, hour, min, sec, nsec, z), new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc)), null];
		}
		if (!((zoneOffset === -1))) {
			t = (_struct$1 = Date(year, (month >> 0), day, hour, min, sec, nsec, go$pkg.UTC), new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
			t.sec = (x$1 = t.sec, x$2 = new Go$Int64(0, zoneOffset), new Go$Int64(x$1.high - x$2.high, x$1.low - x$2.low));
			_tuple$22 = local.lookup((x$3 = t.sec, new Go$Int64(x$3.high + -15, x$3.low + 2288912640))); name = _tuple$22[0]; offset = _tuple$22[1];
			if ((offset === zoneOffset) && (zoneName === "" || name === zoneName)) {
				t.loc = local;
				return [(_struct$2 = t, new Time.Ptr(_struct$2.sec, _struct$2.nsec, _struct$2.loc)), null];
			}
			t.loc = FixedZone(zoneName, zoneOffset);
			return [(_struct$3 = t, new Time.Ptr(_struct$3.sec, _struct$3.nsec, _struct$3.loc)), null];
		}
		if (!(zoneName === "")) {
			t$1 = (_struct$4 = Date(year, (month >> 0), day, hour, min, sec, nsec, go$pkg.UTC), new Time.Ptr(_struct$4.sec, _struct$4.nsec, _struct$4.loc));
			_tuple$23 = local.lookupName(zoneName, (x$4 = t$1.sec, new Go$Int64(x$4.high + -15, x$4.low + 2288912640))); offset$1 = _tuple$23[0]; ok$1 = _tuple$23[2];
			if (ok$1) {
				t$1.sec = (x$5 = t$1.sec, x$6 = new Go$Int64(0, offset$1), new Go$Int64(x$5.high - x$6.high, x$5.low - x$6.low));
				t$1.loc = local;
				return [(_struct$5 = t$1, new Time.Ptr(_struct$5.sec, _struct$5.nsec, _struct$5.loc)), null];
			}
			if (zoneName.length > 3 && zoneName.substring(0, 3) === "GMT") {
				_tuple$24 = atoi(zoneName.substring(3)); offset$1 = _tuple$24[0];
				offset$1 = (((offset$1 >>> 16 << 16) * 3600 >> 0) + (offset$1 << 16 >>> 16) * 3600) >> 0;
			}
			t$1.loc = FixedZone(zoneName, offset$1);
			return [(_struct$6 = t$1, new Time.Ptr(_struct$6.sec, _struct$6.nsec, _struct$6.loc)), null];
		}
		return [(_struct$7 = Date(year, (month >> 0), day, hour, min, sec, nsec, defaultLocation), new Time.Ptr(_struct$7.sec, _struct$7.nsec, _struct$7.loc)), null];
	};
	parseTimeZone = function(value) {
		var length, ok, _tmp, _tmp$1, _tmp$2, _tmp$3, _tmp$4, _tmp$5, nUpper, c, _ref, _tmp$6, _tmp$7, _tmp$8, _tmp$9, _tmp$10, _tmp$11, _tmp$12, _tmp$13, _tmp$14, _tmp$15;
		length = 0;
		ok = false;
		if (value.length < 3) {
			_tmp = 0; _tmp$1 = false; length = _tmp; ok = _tmp$1;
			return [length, ok];
		}
		if (value.length >= 4 && value.substring(0, 4) === "ChST") {
			_tmp$2 = 4; _tmp$3 = true; length = _tmp$2; ok = _tmp$3;
			return [length, ok];
		}
		if (value.substring(0, 3) === "GMT") {
			length = parseGMT(value);
			_tmp$4 = length; _tmp$5 = true; length = _tmp$4; ok = _tmp$5;
			return [length, ok];
		}
		nUpper = 0;
		nUpper = 0;
		while (nUpper < 6) {
			if (nUpper >= value.length) {
				break;
			}
			c = value.charCodeAt(nUpper);
			if (c < 65 || 90 < c) {
				break;
			}
			nUpper = nUpper + 1 >> 0;
		}
		_ref = nUpper;
		if (_ref === 0 || _ref === 1 || _ref === 2 || _ref === 6) {
			_tmp$6 = 0; _tmp$7 = false; length = _tmp$6; ok = _tmp$7;
			return [length, ok];
		} else if (_ref === 5) {
			if (value.charCodeAt(4) === 84) {
				_tmp$8 = 5; _tmp$9 = true; length = _tmp$8; ok = _tmp$9;
				return [length, ok];
			}
		} else if (_ref === 4) {
			if (value.charCodeAt(3) === 84) {
				_tmp$10 = 4; _tmp$11 = true; length = _tmp$10; ok = _tmp$11;
				return [length, ok];
			}
		} else if (_ref === 3) {
			_tmp$12 = 3; _tmp$13 = true; length = _tmp$12; ok = _tmp$13;
			return [length, ok];
		}
		_tmp$14 = 0; _tmp$15 = false; length = _tmp$14; ok = _tmp$15;
		return [length, ok];
	};
	parseGMT = function(value) {
		var sign, _tuple, x, rem, err;
		value = value.substring(3);
		if (value.length === 0) {
			return 3;
		}
		sign = value.charCodeAt(0);
		if (!((sign === 45)) && !((sign === 43))) {
			return 3;
		}
		_tuple = leadingInt(value.substring(1)); x = _tuple[0]; rem = _tuple[1]; err = _tuple[2];
		if (!(go$interfaceIsEqual(err, null))) {
			return 3;
		}
		if (sign === 45) {
			x = new Go$Int64(-x.high, -x.low);
		}
		if ((x.high === 0 && x.low === 0) || (x.high < -1 || (x.high === -1 && x.low < 4294967282)) || (0 < x.high || (0 === x.high && 12 < x.low))) {
			return 3;
		}
		return (3 + value.length >> 0) - rem.length >> 0;
	};
	parseNanoseconds = function(value, nbytes) {
		var ns, rangeErrString, err, _tuple, scaleDigits, i;
		ns = 0;
		rangeErrString = "";
		err = null;
		if (!((value.charCodeAt(0) === 46))) {
			err = errBad;
			return [ns, rangeErrString, err];
		}
		_tuple = atoi(value.substring(1, nbytes)); ns = _tuple[0]; err = _tuple[1];
		if (!(go$interfaceIsEqual(err, null))) {
			return [ns, rangeErrString, err];
		}
		if (ns < 0 || 1000000000 <= ns) {
			rangeErrString = "fractional second";
			return [ns, rangeErrString, err];
		}
		scaleDigits = 10 - nbytes >> 0;
		i = 0;
		while (i < scaleDigits) {
			ns = (((ns >>> 16 << 16) * 10 >> 0) + (ns << 16 >>> 16) * 10) >> 0;
			i = i + 1 >> 0;
		}
		return [ns, rangeErrString, err];
	};
	leadingInt = function(s) {
		var x, rem, err, i, c, _tmp, _tmp$1, _tmp$2, x$1, x$2, x$3, _tmp$3, _tmp$4, _tmp$5;
		x = new Go$Int64(0, 0);
		rem = "";
		err = null;
		i = 0;
		while (i < s.length) {
			c = s.charCodeAt(i);
			if (c < 48 || c > 57) {
				break;
			}
			if ((x.high > 214748364 || (x.high === 214748364 && x.low >= 3435973835))) {
				_tmp = new Go$Int64(0, 0); _tmp$1 = ""; _tmp$2 = errLeadingInt; x = _tmp; rem = _tmp$1; err = _tmp$2;
				return [x, rem, err];
			}
			x = (x$1 = (x$2 = go$mul64(x, new Go$Int64(0, 10)), x$3 = new Go$Int64(0, c), new Go$Int64(x$2.high + x$3.high, x$2.low + x$3.low)), new Go$Int64(x$1.high - 0, x$1.low - 48));
			i = i + 1 >> 0;
		}
		_tmp$3 = x; _tmp$4 = s.substring(i); _tmp$5 = null; x = _tmp$3; rem = _tmp$4; err = _tmp$5;
		return [x, rem, err];
	};
	readFile = function(name) {
		var _tuple, f, err, buf, ret, n, _tuple$1;
		var go$deferred = [];
		try {
			_tuple = syscall.Open(name, 0, 0); f = _tuple[0]; err = _tuple[1];
			if (!(go$interfaceIsEqual(err, null))) {
				return [(go$sliceType(Go$Uint8)).nil, err];
			}
			go$deferred.push({ recv: syscall, method: "Close", args: [f] });
			buf = go$makeNativeArray("Uint8", 4096, function() { return 0; });
			ret = (go$sliceType(Go$Uint8)).nil;
			n = 0;
			while (true) {
				_tuple$1 = syscall.Read(f, new (go$sliceType(Go$Uint8))(buf)); n = _tuple$1[0]; err = _tuple$1[1];
				if (n > 0) {
					ret = go$appendSlice(ret, go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, n));
				}
				if ((n === 0) || !(go$interfaceIsEqual(err, null))) {
					break;
				}
			}
			return [ret, err];
		} catch(go$err) {
			go$pushErr(go$err);
			return [(go$sliceType(Go$Uint8)).nil, null];
		} finally {
			go$callDeferred(go$deferred);
		}
	};
	open = function(name) {
		var _tuple, fd, err;
		_tuple = syscall.Open(name, 0, 0); fd = _tuple[0]; err = _tuple[1];
		if (!(go$interfaceIsEqual(err, null))) {
			return [0, err];
		}
		return [(fd >>> 0), null];
	};
	closefd = function(fd) {
		syscall.Close((fd >> 0));
	};
	preadn = function(fd, buf, off) {
		var whence, _tuple, err, _tuple$1, m, err$1;
		whence = 0;
		if (off < 0) {
			whence = 2;
		}
		_tuple = syscall.Seek((fd >> 0), new Go$Int64(0, off), whence); err = _tuple[1];
		if (!(go$interfaceIsEqual(err, null))) {
			return err;
		}
		while (buf.length > 0) {
			_tuple$1 = syscall.Read((fd >> 0), buf); m = _tuple$1[0]; err$1 = _tuple$1[1];
			if (m <= 0) {
				if (go$interfaceIsEqual(err$1, null)) {
					return errors.New("short read");
				}
				return err$1;
			}
			buf = go$subslice(buf, m);
		}
		return null;
	};
	Time.Ptr.prototype.After = function(u) {
		var _struct, t, x, x$1, x$2, x$3;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (x = t.sec, x$1 = u.sec, (x.high > x$1.high || (x.high === x$1.high && x.low > x$1.low))) || (x$2 = t.sec, x$3 = u.sec, (x$2.high === x$3.high && x$2.low === x$3.low)) && t.nsec > u.nsec;
	};
	Time.prototype.After = function(u) { return this.go$val.After(u); };
	Time.Ptr.prototype.Before = function(u) {
		var _struct, t, x, x$1, x$2, x$3;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (x = t.sec, x$1 = u.sec, (x.high < x$1.high || (x.high === x$1.high && x.low < x$1.low))) || (x$2 = t.sec, x$3 = u.sec, (x$2.high === x$3.high && x$2.low === x$3.low)) && t.nsec < u.nsec;
	};
	Time.prototype.Before = function(u) { return this.go$val.Before(u); };
	Time.Ptr.prototype.Equal = function(u) {
		var _struct, t, x, x$1;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (x = t.sec, x$1 = u.sec, (x.high === x$1.high && x.low === x$1.low)) && (t.nsec === u.nsec);
	};
	Time.prototype.Equal = function(u) { return this.go$val.Equal(u); };
	Month.prototype.String = function() {
		var m;
		m = this.go$val;
		return months[(m - 1 >> 0)];
	};
	go$ptrType(Month).prototype.String = function() { return new Month(this.go$get()).String(); };
	Weekday.prototype.String = function() {
		var d;
		d = this.go$val;
		return days[d];
	};
	go$ptrType(Weekday).prototype.String = function() { return new Weekday(this.go$get()).String(); };
	Time.Ptr.prototype.IsZero = function() {
		var _struct, t, x;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (x = t.sec, (x.high === 0 && x.low === 0)) && (t.nsec === 0);
	};
	Time.prototype.IsZero = function() { return this.go$val.IsZero(); };
	Time.Ptr.prototype.abs = function() {
		var _struct, t, l, x, sec, x$1, x$2, x$3, _tuple, offset, x$4, x$5;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		l = t.loc;
		if (l === (go$ptrType(Location)).nil || l === localLoc) {
			l = l.get();
		}
		sec = (x = t.sec, new Go$Int64(x.high + -15, x.low + 2288912640));
		if (!(l === utcLoc)) {
			if (!(l.cacheZone === (go$ptrType(zone)).nil) && (x$1 = l.cacheStart, (x$1.high < sec.high || (x$1.high === sec.high && x$1.low <= sec.low))) && (x$2 = l.cacheEnd, (sec.high < x$2.high || (sec.high === x$2.high && sec.low < x$2.low)))) {
				sec = (x$3 = new Go$Int64(0, l.cacheZone.offset), new Go$Int64(sec.high + x$3.high, sec.low + x$3.low));
			} else {
				_tuple = l.lookup(sec); offset = _tuple[1];
				sec = (x$4 = new Go$Int64(0, offset), new Go$Int64(sec.high + x$4.high, sec.low + x$4.low));
			}
		}
		return (x$5 = new Go$Int64(sec.high + 2147483646, sec.low + 450480384), new Go$Uint64(x$5.high, x$5.low));
	};
	Time.prototype.abs = function() { return this.go$val.abs(); };
	Time.Ptr.prototype.locabs = function() {
		var name, offset, abs, _struct, t, l, x, sec, x$1, x$2, _tuple, x$3, x$4;
		name = "";
		offset = 0;
		abs = new Go$Uint64(0, 0);
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		l = t.loc;
		if (l === (go$ptrType(Location)).nil || l === localLoc) {
			l = l.get();
		}
		sec = (x = t.sec, new Go$Int64(x.high + -15, x.low + 2288912640));
		if (!(l === utcLoc)) {
			if (!(l.cacheZone === (go$ptrType(zone)).nil) && (x$1 = l.cacheStart, (x$1.high < sec.high || (x$1.high === sec.high && x$1.low <= sec.low))) && (x$2 = l.cacheEnd, (sec.high < x$2.high || (sec.high === x$2.high && sec.low < x$2.low)))) {
				name = l.cacheZone.name;
				offset = l.cacheZone.offset;
			} else {
				_tuple = l.lookup(sec); name = _tuple[0]; offset = _tuple[1];
			}
			sec = (x$3 = new Go$Int64(0, offset), new Go$Int64(sec.high + x$3.high, sec.low + x$3.low));
		} else {
			name = "UTC";
		}
		abs = (x$4 = new Go$Int64(sec.high + 2147483646, sec.low + 450480384), new Go$Uint64(x$4.high, x$4.low));
		return [name, offset, abs];
	};
	Time.prototype.locabs = function() { return this.go$val.locabs(); };
	Time.Ptr.prototype.Date = function() {
		var year, month, day, _struct, t, _tuple;
		year = 0;
		month = 0;
		day = 0;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.date(true); year = _tuple[0]; month = _tuple[1]; day = _tuple[2];
		return [year, month, day];
	};
	Time.prototype.Date = function() { return this.go$val.Date(); };
	Time.Ptr.prototype.Year = function() {
		var _struct, t, _tuple, year;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.date(false); year = _tuple[0];
		return year;
	};
	Time.prototype.Year = function() { return this.go$val.Year(); };
	Time.Ptr.prototype.Month = function() {
		var _struct, t, _tuple, month;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.date(true); month = _tuple[1];
		return month;
	};
	Time.prototype.Month = function() { return this.go$val.Month(); };
	Time.Ptr.prototype.Day = function() {
		var _struct, t, _tuple, day;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.date(true); day = _tuple[2];
		return day;
	};
	Time.prototype.Day = function() { return this.go$val.Day(); };
	Time.Ptr.prototype.Weekday = function() {
		var _struct, t;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return absWeekday(t.abs());
	};
	Time.prototype.Weekday = function() { return this.go$val.Weekday(); };
	absWeekday = function(abs) {
		var sec, _q;
		sec = go$div64((new Go$Uint64(abs.high + 0, abs.low + 86400)), new Go$Uint64(0, 604800), true);
		return ((_q = (sec.low >> 0) / 86400, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0);
	};
	Time.Ptr.prototype.ISOWeek = function() {
		var year, week, _struct, t, _tuple, month, day, yday, _r, wday, _q, _r$1, jan1wday, _r$2, dec31wday;
		year = 0;
		week = 0;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.date(true); year = _tuple[0]; month = _tuple[1]; day = _tuple[2]; yday = _tuple[3];
		wday = (_r = ((t.Weekday() + 6 >> 0) >> 0) % 7, _r === _r ? _r : go$throwRuntimeError("integer divide by zero"));
		week = (_q = (((yday - wday >> 0) + 7 >> 0)) / 7, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		jan1wday = (_r$1 = (((wday - yday >> 0) + 371 >> 0)) % 7, _r$1 === _r$1 ? _r$1 : go$throwRuntimeError("integer divide by zero"));
		if (1 <= jan1wday && jan1wday <= 3) {
			week = week + 1 >> 0;
		}
		if (week === 0) {
			year = year - 1 >> 0;
			week = 52;
			if ((jan1wday === 4) || ((jan1wday === 5) && isLeap(year))) {
				week = week + 1 >> 0;
			}
		}
		if ((month === 12) && day >= 29 && wday < 3) {
			dec31wday = (_r$2 = (((wday + 31 >> 0) - day >> 0)) % 7, _r$2 === _r$2 ? _r$2 : go$throwRuntimeError("integer divide by zero"));
			if (0 <= dec31wday && dec31wday <= 2) {
				year = year + 1 >> 0;
				week = 1;
			}
		}
		return [year, week];
	};
	Time.prototype.ISOWeek = function() { return this.go$val.ISOWeek(); };
	Time.Ptr.prototype.Clock = function() {
		var hour, min, sec, _struct, t, _tuple;
		hour = 0;
		min = 0;
		sec = 0;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = absClock(t.abs()); hour = _tuple[0]; min = _tuple[1]; sec = _tuple[2];
		return [hour, min, sec];
	};
	Time.prototype.Clock = function() { return this.go$val.Clock(); };
	absClock = function(abs) {
		var hour, min, sec, _q, _q$1;
		hour = 0;
		min = 0;
		sec = 0;
		sec = (go$div64(abs, new Go$Uint64(0, 86400), true).low >> 0);
		hour = (_q = sec / 3600, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		sec = sec - (((((hour >>> 16 << 16) * 3600 >> 0) + (hour << 16 >>> 16) * 3600) >> 0)) >> 0;
		min = (_q$1 = sec / 60, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >> 0 : go$throwRuntimeError("integer divide by zero"));
		sec = sec - (((((min >>> 16 << 16) * 60 >> 0) + (min << 16 >>> 16) * 60) >> 0)) >> 0;
		return [hour, min, sec];
	};
	Time.Ptr.prototype.Hour = function() {
		var _struct, t, _q;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (_q = (go$div64(t.abs(), new Go$Uint64(0, 86400), true).low >> 0) / 3600, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
	};
	Time.prototype.Hour = function() { return this.go$val.Hour(); };
	Time.Ptr.prototype.Minute = function() {
		var _struct, t, _q;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (_q = (go$div64(t.abs(), new Go$Uint64(0, 3600), true).low >> 0) / 60, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
	};
	Time.prototype.Minute = function() { return this.go$val.Minute(); };
	Time.Ptr.prototype.Second = function() {
		var _struct, t;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (go$div64(t.abs(), new Go$Uint64(0, 60), true).low >> 0);
	};
	Time.prototype.Second = function() { return this.go$val.Second(); };
	Time.Ptr.prototype.Nanosecond = function() {
		var _struct, t;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (t.nsec >> 0);
	};
	Time.prototype.Nanosecond = function() { return this.go$val.Nanosecond(); };
	Time.Ptr.prototype.YearDay = function() {
		var _struct, t, _tuple, yday;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.date(false); yday = _tuple[3];
		return yday + 1 >> 0;
	};
	Time.prototype.YearDay = function() { return this.go$val.YearDay(); };
	Duration.prototype.String = function() {
		var d, buf, w, u, neg, prec, unit, _tuple, _tuple$1;
		d = this;
		buf = go$makeNativeArray("Uint8", 32, function() { return 0; });
		w = 32;
		u = new Go$Uint64(d.high, d.low);
		neg = (d.high < 0 || (d.high === 0 && d.low < 0));
		if (neg) {
			u = new Go$Uint64(-u.high, -u.low);
		}
		if ((u.high < 0 || (u.high === 0 && u.low < 1000000000))) {
			prec = 0;
			unit = 0;
			if ((u.high === 0 && u.low === 0)) {
				return "0";
			} else if ((u.high < 0 || (u.high === 0 && u.low < 1000))) {
				prec = 0;
				unit = 110;
			} else if ((u.high < 0 || (u.high === 0 && u.low < 1000000))) {
				prec = 3;
				unit = 117;
			} else {
				prec = 6;
				unit = 109;
			}
			w = w - 2 >> 0;
			buf[w] = unit;
			buf[(w + 1 >> 0)] = 115;
			_tuple = fmtFrac(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w), u, prec); w = _tuple[0]; u = _tuple[1];
			w = fmtInt(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w), u);
		} else {
			w = w - 1 >> 0;
			buf[w] = 115;
			_tuple$1 = fmtFrac(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w), u, 9); w = _tuple$1[0]; u = _tuple$1[1];
			w = fmtInt(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w), go$div64(u, new Go$Uint64(0, 60), true));
			u = go$div64(u, new Go$Uint64(0, 60), false);
			if ((u.high > 0 || (u.high === 0 && u.low > 0))) {
				w = w - 1 >> 0;
				buf[w] = 109;
				w = fmtInt(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w), go$div64(u, new Go$Uint64(0, 60), true));
				u = go$div64(u, new Go$Uint64(0, 60), false);
				if ((u.high > 0 || (u.high === 0 && u.low > 0))) {
					w = w - 1 >> 0;
					buf[w] = 104;
					w = fmtInt(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w), u);
				}
			}
		}
		if (neg) {
			w = w - 1 >> 0;
			buf[w] = 45;
		}
		return go$bytesToString(go$subslice(new (go$sliceType(Go$Uint8))(buf), w));
	};
	go$ptrType(Duration).prototype.String = function() { return this.go$get().String(); };
	fmtFrac = function(buf, v, prec) {
		var nw, nv, w, print, i, digit, _tmp, _tmp$1;
		nw = 0;
		nv = new Go$Uint64(0, 0);
		w = buf.length;
		print = false;
		i = 0;
		while (i < prec) {
			digit = go$div64(v, new Go$Uint64(0, 10), true);
			print = print || !((digit.high === 0 && digit.low === 0));
			if (print) {
				w = w - 1 >> 0;
				(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = (digit.low << 24 >>> 24) + 48 << 24 >>> 24;
			}
			v = go$div64(v, new Go$Uint64(0, 10), false);
			i = i + 1 >> 0;
		}
		if (print) {
			w = w - 1 >> 0;
			(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = 46;
		}
		_tmp = w; _tmp$1 = v; nw = _tmp; nv = _tmp$1;
		return [nw, nv];
	};
	fmtInt = function(buf, v) {
		var w;
		w = buf.length;
		if ((v.high === 0 && v.low === 0)) {
			w = w - 1 >> 0;
			(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = 48;
		} else {
			while ((v.high > 0 || (v.high === 0 && v.low > 0))) {
				w = w - 1 >> 0;
				(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = (go$div64(v, new Go$Uint64(0, 10), true).low << 24 >>> 24) + 48 << 24 >>> 24;
				v = go$div64(v, new Go$Uint64(0, 10), false);
			}
		}
		return w;
	};
	Duration.prototype.Nanoseconds = function() {
		var d;
		d = this;
		return new Go$Int64(d.high, d.low);
	};
	go$ptrType(Duration).prototype.Nanoseconds = function() { return this.go$get().Nanoseconds(); };
	Duration.prototype.Seconds = function() {
		var d, sec, nsec;
		d = this;
		sec = go$div64(d, new Duration(0, 1000000000), false);
		nsec = go$div64(d, new Duration(0, 1000000000), true);
		return go$flatten64(sec) + go$flatten64(nsec) * 1e-09;
	};
	go$ptrType(Duration).prototype.Seconds = function() { return this.go$get().Seconds(); };
	Duration.prototype.Minutes = function() {
		var d, min, nsec;
		d = this;
		min = go$div64(d, new Duration(13, 4165425152), false);
		nsec = go$div64(d, new Duration(13, 4165425152), true);
		return go$flatten64(min) + go$flatten64(nsec) * 1.6666666666666667e-11;
	};
	go$ptrType(Duration).prototype.Minutes = function() { return this.go$get().Minutes(); };
	Duration.prototype.Hours = function() {
		var d, hour, nsec;
		d = this;
		hour = go$div64(d, new Duration(838, 817405952), false);
		nsec = go$div64(d, new Duration(838, 817405952), true);
		return go$flatten64(hour) + go$flatten64(nsec) * 2.777777777777778e-13;
	};
	go$ptrType(Duration).prototype.Hours = function() { return this.go$get().Hours(); };
	Time.Ptr.prototype.Add = function(d) {
		var _struct, t, x, x$1, x$2, x$3, nsec, x$4, x$5, _struct$1;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		t.sec = (x = t.sec, x$1 = (x$2 = go$div64(d, new Duration(0, 1000000000), false), new Go$Int64(x$2.high, x$2.low)), new Go$Int64(x.high + x$1.high, x.low + x$1.low));
		nsec = (t.nsec >> 0) + ((x$3 = go$div64(d, new Duration(0, 1000000000), true), x$3.low + ((x$3.high >> 31) * 4294967296)) >> 0) >> 0;
		if (nsec >= 1000000000) {
			t.sec = (x$4 = t.sec, new Go$Int64(x$4.high + 0, x$4.low + 1));
			nsec = nsec - 1000000000 >> 0;
		} else if (nsec < 0) {
			t.sec = (x$5 = t.sec, new Go$Int64(x$5.high - 0, x$5.low - 1));
			nsec = nsec + 1000000000 >> 0;
		}
		t.nsec = (nsec >>> 0);
		return (_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
	};
	Time.prototype.Add = function(d) { return this.go$val.Add(d); };
	Time.Ptr.prototype.Sub = function(u) {
		var _struct, t, x, x$1, x$2, x$3, x$4, d, _struct$1, _struct$2;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		d = (x = go$mul64((x$1 = (x$2 = t.sec, x$3 = u.sec, new Go$Int64(x$2.high - x$3.high, x$2.low - x$3.low)), new Duration(x$1.high, x$1.low)), new Duration(0, 1000000000)), x$4 = new Duration(0, ((t.nsec >> 0) - (u.nsec >> 0) >> 0)), new Duration(x.high + x$4.high, x.low + x$4.low));
		if (u.Add(d).Equal((_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc)))) {
			return d;
		} else if (t.Before((_struct$2 = u, new Time.Ptr(_struct$2.sec, _struct$2.nsec, _struct$2.loc)))) {
			return new Duration(-2147483648, 0);
		} else {
			return new Duration(2147483647, 4294967295);
		}
	};
	Time.prototype.Sub = function(u) { return this.go$val.Sub(u); };
	Time.Ptr.prototype.AddDate = function(years, months$1, days$1) {
		var _struct, t, _tuple, year, month, day, _tuple$1, hour, min, sec, _struct$1;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.Date(); year = _tuple[0]; month = _tuple[1]; day = _tuple[2];
		_tuple$1 = t.Clock(); hour = _tuple$1[0]; min = _tuple$1[1]; sec = _tuple$1[2];
		return (_struct$1 = Date(year + years >> 0, month + (months$1 >> 0) >> 0, day + days$1 >> 0, hour, min, sec, (t.nsec >> 0), t.loc), new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
	};
	Time.prototype.AddDate = function(years, months$1, days$1) { return this.go$val.AddDate(years, months$1, days$1); };
	Time.Ptr.prototype.date = function(full) {
		var year, month, day, yday, _struct, t, _tuple;
		year = 0;
		month = 0;
		day = 0;
		yday = 0;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = absDate(t.abs(), full); year = _tuple[0]; month = _tuple[1]; day = _tuple[2]; yday = _tuple[3];
		return [year, month, day, yday];
	};
	Time.prototype.date = function(full) { return this.go$val.date(full); };
	absDate = function(abs, full) {
		var year, month, day, yday, d, n, y, x, x$1, x$2, x$3, x$4, x$5, x$6, x$7, x$8, x$9, x$10, _q, end, begin;
		year = 0;
		month = 0;
		day = 0;
		yday = 0;
		d = go$div64(abs, new Go$Uint64(0, 86400), false);
		n = go$div64(d, new Go$Uint64(0, 146097), false);
		y = go$mul64(new Go$Uint64(0, 400), n);
		d = (x = go$mul64(new Go$Uint64(0, 146097), n), new Go$Uint64(d.high - x.high, d.low - x.low));
		n = go$div64(d, new Go$Uint64(0, 36524), false);
		n = (x$1 = go$shiftRightUint64(n, 2), new Go$Uint64(n.high - x$1.high, n.low - x$1.low));
		y = (x$2 = go$mul64(new Go$Uint64(0, 100), n), new Go$Uint64(y.high + x$2.high, y.low + x$2.low));
		d = (x$3 = go$mul64(new Go$Uint64(0, 36524), n), new Go$Uint64(d.high - x$3.high, d.low - x$3.low));
		n = go$div64(d, new Go$Uint64(0, 1461), false);
		y = (x$4 = go$mul64(new Go$Uint64(0, 4), n), new Go$Uint64(y.high + x$4.high, y.low + x$4.low));
		d = (x$5 = go$mul64(new Go$Uint64(0, 1461), n), new Go$Uint64(d.high - x$5.high, d.low - x$5.low));
		n = go$div64(d, new Go$Uint64(0, 365), false);
		n = (x$6 = go$shiftRightUint64(n, 2), new Go$Uint64(n.high - x$6.high, n.low - x$6.low));
		y = (x$7 = n, new Go$Uint64(y.high + x$7.high, y.low + x$7.low));
		d = (x$8 = go$mul64(new Go$Uint64(0, 365), n), new Go$Uint64(d.high - x$8.high, d.low - x$8.low));
		year = ((x$9 = (x$10 = new Go$Int64(y.high, y.low), new Go$Int64(x$10.high + -69, x$10.low + 4075721025)), x$9.low + ((x$9.high >> 31) * 4294967296)) >> 0);
		yday = (d.low >> 0);
		if (!full) {
			return [year, month, day, yday];
		}
		day = yday;
		if (isLeap(year)) {
			if (day > 59) {
				day = day - 1 >> 0;
			} else if (day === 59) {
				month = 2;
				day = 29;
				return [year, month, day, yday];
			}
		}
		month = ((_q = day / 31, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0);
		end = (daysBefore[(month + 1 >> 0)] >> 0);
		begin = 0;
		if (day >= end) {
			month = month + 1 >> 0;
			begin = end;
		} else {
			begin = (daysBefore[month] >> 0);
		}
		month = month + 1 >> 0;
		day = (day - begin >> 0) + 1 >> 0;
		return [year, month, day, yday];
	};
	Time.Ptr.prototype.UTC = function() {
		var _struct, t, _struct$1;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		t.loc = go$pkg.UTC;
		return (_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
	};
	Time.prototype.UTC = function() { return this.go$val.UTC(); };
	Time.Ptr.prototype.Local = function() {
		var _struct, t, _struct$1;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		t.loc = go$pkg.Local;
		return (_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
	};
	Time.prototype.Local = function() { return this.go$val.Local(); };
	Time.Ptr.prototype.In = function(loc) {
		var _struct, t, _struct$1;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		if (loc === (go$ptrType(Location)).nil) {
			throw go$panic(new Go$String("time: missing Location in call to Time.In"));
		}
		t.loc = loc;
		return (_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
	};
	Time.prototype.In = function(loc) { return this.go$val.In(loc); };
	Time.Ptr.prototype.Location = function() {
		var _struct, t, l;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		l = t.loc;
		if (l === (go$ptrType(Location)).nil) {
			l = go$pkg.UTC;
		}
		return l;
	};
	Time.prototype.Location = function() { return this.go$val.Location(); };
	Time.Ptr.prototype.Zone = function() {
		var name, offset, _struct, t, _tuple, x;
		name = "";
		offset = 0;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		_tuple = t.loc.lookup((x = t.sec, new Go$Int64(x.high + -15, x.low + 2288912640))); name = _tuple[0]; offset = _tuple[1];
		return [name, offset];
	};
	Time.prototype.Zone = function() { return this.go$val.Zone(); };
	Time.Ptr.prototype.Unix = function() {
		var _struct, t, x;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (x = t.sec, new Go$Int64(x.high + -15, x.low + 2288912640));
	};
	Time.prototype.Unix = function() { return this.go$val.Unix(); };
	Time.Ptr.prototype.UnixNano = function() {
		var _struct, t, x, x$1, x$2, x$3;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return (x = go$mul64(((x$1 = t.sec, new Go$Int64(x$1.high + -15, x$1.low + 2288912640))), new Go$Int64(0, 1000000000)), x$2 = (x$3 = t.nsec, new Go$Int64(0, x$3.constructor === Number ? x$3 : 1)), new Go$Int64(x.high + x$2.high, x.low + x$2.low));
	};
	Time.prototype.UnixNano = function() { return this.go$val.UnixNano(); };
	Time.Ptr.prototype.MarshalBinary = function() {
		var _struct, t, offsetMin, _tuple, offset, _r, _q, enc;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		offsetMin = 0;
		if (t.Location() === utcLoc) {
			offsetMin = -1;
		} else {
			_tuple = t.Zone(); offset = _tuple[1];
			if (!(((_r = offset % 60, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) === 0))) {
				return [(go$sliceType(Go$Uint8)).nil, errors.New("Time.MarshalBinary: zone offset has fractional minute")];
			}
			offset = (_q = offset / 60, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
			if (offset < -32768 || (offset === -1) || offset > 32767) {
				return [(go$sliceType(Go$Uint8)).nil, errors.New("Time.MarshalBinary: unexpected zone offset")];
			}
			offsetMin = (offset << 16 >> 16);
		}
		enc = new (go$sliceType(Go$Uint8))([1, (go$shiftRightInt64(t.sec, 56).low << 24 >>> 24), (go$shiftRightInt64(t.sec, 48).low << 24 >>> 24), (go$shiftRightInt64(t.sec, 40).low << 24 >>> 24), (go$shiftRightInt64(t.sec, 32).low << 24 >>> 24), (go$shiftRightInt64(t.sec, 24).low << 24 >>> 24), (go$shiftRightInt64(t.sec, 16).low << 24 >>> 24), (go$shiftRightInt64(t.sec, 8).low << 24 >>> 24), (t.sec.low << 24 >>> 24), ((t.nsec >>> 24 >>> 0) << 24 >>> 24), ((t.nsec >>> 16 >>> 0) << 24 >>> 24), ((t.nsec >>> 8 >>> 0) << 24 >>> 24), (t.nsec << 24 >>> 24), ((offsetMin >> 8 << 16 >> 16) << 24 >>> 24), (offsetMin << 24 >>> 24)]);
		return [enc, null];
	};
	Time.prototype.MarshalBinary = function() { return this.go$val.MarshalBinary(); };
	Time.Ptr.prototype.UnmarshalBinary = function(data$1) {
		var t, buf, x, x$1, x$2, x$3, x$4, x$5, x$6, x$7, x$8, x$9, x$10, x$11, x$12, x$13, x$14, offset, _tuple, x$15, localoff;
		t = this;
		buf = data$1;
		if (buf.length === 0) {
			return errors.New("Time.UnmarshalBinary: no data");
		}
		if (!((((0 < 0 || 0 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 0]) === 1))) {
			return errors.New("Time.UnmarshalBinary: unsupported version");
		}
		if (!((buf.length === 15))) {
			return errors.New("Time.UnmarshalBinary: invalid length");
		}
		buf = go$subslice(buf, 1);
		t.sec = (x = (x$1 = (x$2 = (x$3 = (x$4 = (x$5 = (x$6 = new Go$Int64(0, ((7 < 0 || 7 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 7])), x$7 = go$shiftLeft64(new Go$Int64(0, ((6 < 0 || 6 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 6])), 8), new Go$Int64(x$6.high | x$7.high, (x$6.low | x$7.low) >>> 0)), x$8 = go$shiftLeft64(new Go$Int64(0, ((5 < 0 || 5 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 5])), 16), new Go$Int64(x$5.high | x$8.high, (x$5.low | x$8.low) >>> 0)), x$9 = go$shiftLeft64(new Go$Int64(0, ((4 < 0 || 4 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 4])), 24), new Go$Int64(x$4.high | x$9.high, (x$4.low | x$9.low) >>> 0)), x$10 = go$shiftLeft64(new Go$Int64(0, ((3 < 0 || 3 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 3])), 32), new Go$Int64(x$3.high | x$10.high, (x$3.low | x$10.low) >>> 0)), x$11 = go$shiftLeft64(new Go$Int64(0, ((2 < 0 || 2 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 2])), 40), new Go$Int64(x$2.high | x$11.high, (x$2.low | x$11.low) >>> 0)), x$12 = go$shiftLeft64(new Go$Int64(0, ((1 < 0 || 1 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 1])), 48), new Go$Int64(x$1.high | x$12.high, (x$1.low | x$12.low) >>> 0)), x$13 = go$shiftLeft64(new Go$Int64(0, ((0 < 0 || 0 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 0])), 56), new Go$Int64(x.high | x$13.high, (x.low | x$13.low) >>> 0));
		buf = go$subslice(buf, 8);
		t.nsec = (((((((3 < 0 || 3 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 3]) >> 0) | ((((2 < 0 || 2 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 2]) >> 0) << 8 >> 0)) | ((((1 < 0 || 1 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 1]) >> 0) << 16 >> 0)) | ((((0 < 0 || 0 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 0]) >> 0) << 24 >> 0)) >>> 0);
		buf = go$subslice(buf, 4);
		offset = (x$14 = (((((1 < 0 || 1 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 1]) << 16 >> 16) | ((((0 < 0 || 0 >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + 0]) << 16 >> 16) << 8 << 16 >> 16)) >> 0), (((x$14 >>> 16 << 16) * 60 >> 0) + (x$14 << 16 >>> 16) * 60) >> 0);
		if (offset === -60) {
			t.loc = utcLoc;
		} else {
			_tuple = go$pkg.Local.lookup((x$15 = t.sec, new Go$Int64(x$15.high + -15, x$15.low + 2288912640))); localoff = _tuple[1];
			if (offset === localoff) {
				t.loc = go$pkg.Local;
			} else {
				t.loc = FixedZone("", offset);
			}
		}
		return null;
	};
	Time.prototype.UnmarshalBinary = function(data$1) { return this.go$val.UnmarshalBinary(data$1); };
	Time.Ptr.prototype.GobEncode = function() {
		var _struct, t;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		return t.MarshalBinary();
	};
	Time.prototype.GobEncode = function() { return this.go$val.GobEncode(); };
	Time.Ptr.prototype.GobDecode = function(data$1) {
		var t;
		t = this;
		return t.UnmarshalBinary(data$1);
	};
	Time.prototype.GobDecode = function(data$1) { return this.go$val.GobDecode(data$1); };
	Time.Ptr.prototype.MarshalJSON = function() {
		var _struct, t, y;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		y = t.Year();
		if (y < 0 || y >= 10000) {
			return [(go$sliceType(Go$Uint8)).nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")];
		}
		return [new (go$sliceType(Go$Uint8))(go$stringToBytes(t.Format("\"2006-01-02T15:04:05.999999999Z07:00\""))), null];
	};
	Time.prototype.MarshalJSON = function() { return this.go$val.MarshalJSON(); };
	Time.Ptr.prototype.UnmarshalJSON = function(data$1) {
		var err, t, _tuple, _struct, l, r;
		err = null;
		t = this;
		_tuple = Parse("\"2006-01-02T15:04:05Z07:00\"", go$bytesToString(data$1)); l = t; r = (_struct = _tuple[0], new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc)); l.sec = r.sec; l.nsec = r.nsec; l.loc = r.loc; err = _tuple[1];
		return err;
	};
	Time.prototype.UnmarshalJSON = function(data$1) { return this.go$val.UnmarshalJSON(data$1); };
	Time.Ptr.prototype.MarshalText = function() {
		var _struct, t, y;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		y = t.Year();
		if (y < 0 || y >= 10000) {
			return [(go$sliceType(Go$Uint8)).nil, errors.New("Time.MarshalText: year outside of range [0,9999]")];
		}
		return [new (go$sliceType(Go$Uint8))(go$stringToBytes(t.Format("2006-01-02T15:04:05.999999999Z07:00"))), null];
	};
	Time.prototype.MarshalText = function() { return this.go$val.MarshalText(); };
	Time.Ptr.prototype.UnmarshalText = function(data$1) {
		var err, t, _tuple, _struct, l, r;
		err = null;
		t = this;
		_tuple = Parse("2006-01-02T15:04:05Z07:00", go$bytesToString(data$1)); l = t; r = (_struct = _tuple[0], new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc)); l.sec = r.sec; l.nsec = r.nsec; l.loc = r.loc; err = _tuple[1];
		return err;
	};
	Time.prototype.UnmarshalText = function(data$1) { return this.go$val.UnmarshalText(data$1); };
	Unix = go$pkg.Unix = function(sec, nsec) {
		var n, x, x$1;
		if ((nsec.high < 0 || (nsec.high === 0 && nsec.low < 0)) || (nsec.high > 0 || (nsec.high === 0 && nsec.low >= 1000000000))) {
			n = go$div64(nsec, new Go$Int64(0, 1000000000), false);
			sec = (x = n, new Go$Int64(sec.high + x.high, sec.low + x.low));
			nsec = (x$1 = go$mul64(n, new Go$Int64(0, 1000000000)), new Go$Int64(nsec.high - x$1.high, nsec.low - x$1.low));
			if ((nsec.high < 0 || (nsec.high === 0 && nsec.low < 0))) {
				nsec = new Go$Int64(nsec.high + 0, nsec.low + 1000000000);
				sec = new Go$Int64(sec.high - 0, sec.low - 1);
			}
		}
		return new Time.Ptr(new Go$Int64(sec.high + 14, sec.low + 2006054656), (nsec.low >>> 0), go$pkg.Local);
	};
	isLeap = function(year) {
		var _r, _r$1, _r$2;
		return ((_r = year % 4, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) === 0) && (!(((_r$1 = year % 100, _r$1 === _r$1 ? _r$1 : go$throwRuntimeError("integer divide by zero")) === 0)) || ((_r$2 = year % 400, _r$2 === _r$2 ? _r$2 : go$throwRuntimeError("integer divide by zero")) === 0));
	};
	norm = function(hi, lo, base) {
		var nhi, nlo, _q, n, _q$1, n$1, _tmp, _tmp$1;
		nhi = 0;
		nlo = 0;
		if (lo < 0) {
			n = (_q = ((-lo - 1 >> 0)) / base, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) + 1 >> 0;
			hi = hi - (n) >> 0;
			lo = lo + (((((n >>> 16 << 16) * base >> 0) + (n << 16 >>> 16) * base) >> 0)) >> 0;
		}
		if (lo >= base) {
			n$1 = (_q$1 = lo / base, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >> 0 : go$throwRuntimeError("integer divide by zero"));
			hi = hi + (n$1) >> 0;
			lo = lo - (((((n$1 >>> 16 << 16) * base >> 0) + (n$1 << 16 >>> 16) * base) >> 0)) >> 0;
		}
		_tmp = hi; _tmp$1 = lo; nhi = _tmp; nlo = _tmp$1;
		return [nhi, nlo];
	};
	Date = go$pkg.Date = function(year, month, day, hour, min, sec, nsec, loc) {
		var m, _tuple, _tuple$1, _tuple$2, _tuple$3, _tuple$4, x, x$1, y, n, x$2, d, x$3, x$4, x$5, x$6, x$7, x$8, x$9, abs, x$10, x$11, unix, _tuple$5, offset, start, end, x$12, utc, _tuple$6, _tuple$7, x$13;
		if (loc === (go$ptrType(Location)).nil) {
			throw go$panic(new Go$String("time: missing Location in call to Date"));
		}
		m = (month >> 0) - 1 >> 0;
		_tuple = norm(year, m, 12); year = _tuple[0]; m = _tuple[1];
		month = (m >> 0) + 1 >> 0;
		_tuple$1 = norm(sec, nsec, 1000000000); sec = _tuple$1[0]; nsec = _tuple$1[1];
		_tuple$2 = norm(min, sec, 60); min = _tuple$2[0]; sec = _tuple$2[1];
		_tuple$3 = norm(hour, min, 60); hour = _tuple$3[0]; min = _tuple$3[1];
		_tuple$4 = norm(day, hour, 24); day = _tuple$4[0]; hour = _tuple$4[1];
		y = (x = (x$1 = new Go$Int64(0, year), new Go$Int64(x$1.high - -69, x$1.low - 4075721025)), new Go$Uint64(x.high, x.low));
		n = go$div64(y, new Go$Uint64(0, 400), false);
		y = (x$2 = go$mul64(new Go$Uint64(0, 400), n), new Go$Uint64(y.high - x$2.high, y.low - x$2.low));
		d = go$mul64(new Go$Uint64(0, 146097), n);
		n = go$div64(y, new Go$Uint64(0, 100), false);
		y = (x$3 = go$mul64(new Go$Uint64(0, 100), n), new Go$Uint64(y.high - x$3.high, y.low - x$3.low));
		d = (x$4 = go$mul64(new Go$Uint64(0, 36524), n), new Go$Uint64(d.high + x$4.high, d.low + x$4.low));
		n = go$div64(y, new Go$Uint64(0, 4), false);
		y = (x$5 = go$mul64(new Go$Uint64(0, 4), n), new Go$Uint64(y.high - x$5.high, y.low - x$5.low));
		d = (x$6 = go$mul64(new Go$Uint64(0, 1461), n), new Go$Uint64(d.high + x$6.high, d.low + x$6.low));
		n = y;
		d = (x$7 = go$mul64(new Go$Uint64(0, 365), n), new Go$Uint64(d.high + x$7.high, d.low + x$7.low));
		d = (x$8 = new Go$Uint64(0, daysBefore[(month - 1 >> 0)]), new Go$Uint64(d.high + x$8.high, d.low + x$8.low));
		if (isLeap(year) && month >= 3) {
			d = new Go$Uint64(d.high + 0, d.low + 1);
		}
		d = (x$9 = new Go$Uint64(0, (day - 1 >> 0)), new Go$Uint64(d.high + x$9.high, d.low + x$9.low));
		abs = go$mul64(d, new Go$Uint64(0, 86400));
		abs = (x$10 = new Go$Uint64(0, ((((((hour >>> 16 << 16) * 3600 >> 0) + (hour << 16 >>> 16) * 3600) >> 0) + ((((min >>> 16 << 16) * 60 >> 0) + (min << 16 >>> 16) * 60) >> 0) >> 0) + sec >> 0)), new Go$Uint64(abs.high + x$10.high, abs.low + x$10.low));
		unix = (x$11 = new Go$Int64(abs.high, abs.low), new Go$Int64(x$11.high + -2147483647, x$11.low + 3844486912));
		_tuple$5 = loc.lookup(unix); offset = _tuple$5[1]; start = _tuple$5[3]; end = _tuple$5[4];
		if (!((offset === 0))) {
			utc = (x$12 = new Go$Int64(0, offset), new Go$Int64(unix.high - x$12.high, unix.low - x$12.low));
			if ((utc.high < start.high || (utc.high === start.high && utc.low < start.low))) {
				_tuple$6 = loc.lookup(new Go$Int64(start.high - 0, start.low - 1)); offset = _tuple$6[1];
			} else if ((utc.high > end.high || (utc.high === end.high && utc.low >= end.low))) {
				_tuple$7 = loc.lookup(end); offset = _tuple$7[1];
			}
			unix = (x$13 = new Go$Int64(0, offset), new Go$Int64(unix.high - x$13.high, unix.low - x$13.low));
		}
		return new Time.Ptr(new Go$Int64(unix.high + 14, unix.low + 2006054656), (nsec >>> 0), loc);
	};
	Time.Ptr.prototype.Truncate = function(d) {
		var _struct, t, _struct$1, _tuple, _struct$2, r, _struct$3;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		if ((d.high < 0 || (d.high === 0 && d.low <= 0))) {
			return (_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
		}
		_tuple = div((_struct$2 = t, new Time.Ptr(_struct$2.sec, _struct$2.nsec, _struct$2.loc)), d); r = _tuple[1];
		return (_struct$3 = t.Add(new Duration(-r.high, -r.low)), new Time.Ptr(_struct$3.sec, _struct$3.nsec, _struct$3.loc));
	};
	Time.prototype.Truncate = function(d) { return this.go$val.Truncate(d); };
	Time.Ptr.prototype.Round = function(d) {
		var _struct, t, _struct$1, _tuple, _struct$2, r, x, _struct$3, _struct$4;
		t = (_struct = this, new Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
		if ((d.high < 0 || (d.high === 0 && d.low <= 0))) {
			return (_struct$1 = t, new Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc));
		}
		_tuple = div((_struct$2 = t, new Time.Ptr(_struct$2.sec, _struct$2.nsec, _struct$2.loc)), d); r = _tuple[1];
		if ((x = new Duration(r.high + r.high, r.low + r.low), (x.high < d.high || (x.high === d.high && x.low < d.low)))) {
			return (_struct$3 = t.Add(new Duration(-r.high, -r.low)), new Time.Ptr(_struct$3.sec, _struct$3.nsec, _struct$3.loc));
		}
		return (_struct$4 = t.Add(new Duration(d.high - r.high, d.low - r.low)), new Time.Ptr(_struct$4.sec, _struct$4.nsec, _struct$4.loc));
	};
	Time.prototype.Round = function(d) { return this.go$val.Round(d); };
	div = function(t, d) {
		var qmod2, r, neg, nsec, x, x$1, x$2, x$3, x$4, _q, _r, x$5, d1, x$6, x$7, x$8, x$9, x$10, sec, tmp, u1, u0, _tmp, _tmp$1, u0x, _tmp$2, _tmp$3, x$11, d1$1, x$12, d0, _tmp$4, _tmp$5, x$13, x$14, x$15;
		qmod2 = 0;
		r = new Duration(0, 0);
		neg = false;
		nsec = (t.nsec >> 0);
		if ((x = t.sec, (x.high < 0 || (x.high === 0 && x.low < 0)))) {
			neg = true;
			t.sec = (x$1 = t.sec, new Go$Int64(-x$1.high, -x$1.low));
			nsec = -nsec;
			if (nsec < 0) {
				nsec = nsec + 1000000000 >> 0;
				t.sec = (x$2 = t.sec, new Go$Int64(x$2.high - 0, x$2.low - 1));
			}
		}
		if ((d.high < 0 || (d.high === 0 && d.low < 1000000000)) && (x$3 = go$div64(new Duration(0, 1000000000), (new Duration(d.high + d.high, d.low + d.low)), true), (x$3.high === 0 && x$3.low === 0))) {
			qmod2 = ((_q = nsec / ((d.low + ((d.high >> 31) * 4294967296)) >> 0), (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0) & 1;
			r = new Duration(0, (_r = nsec % ((d.low + ((d.high >> 31) * 4294967296)) >> 0), _r === _r ? _r : go$throwRuntimeError("integer divide by zero")));
		} else if ((x$4 = go$div64(d, new Duration(0, 1000000000), true), (x$4.high === 0 && x$4.low === 0))) {
			d1 = (x$5 = go$div64(d, new Duration(0, 1000000000), false), new Go$Int64(x$5.high, x$5.low));
			qmod2 = ((x$6 = go$div64(t.sec, d1, false), x$6.low + ((x$6.high >> 31) * 4294967296)) >> 0) & 1;
			r = (x$7 = go$mul64((x$8 = go$div64(t.sec, d1, true), new Duration(x$8.high, x$8.low)), new Duration(0, 1000000000)), x$9 = new Duration(0, nsec), new Duration(x$7.high + x$9.high, x$7.low + x$9.low));
		} else {
			sec = (x$10 = t.sec, new Go$Uint64(x$10.high, x$10.low));
			tmp = go$mul64((go$shiftRightUint64(sec, 32)), new Go$Uint64(0, 1000000000));
			u1 = go$shiftRightUint64(tmp, 32);
			u0 = go$shiftLeft64(tmp, 32);
			tmp = go$mul64(new Go$Uint64(sec.high & 0, (sec.low & 4294967295) >>> 0), new Go$Uint64(0, 1000000000));
			_tmp = u0; _tmp$1 = new Go$Uint64(u0.high + tmp.high, u0.low + tmp.low); u0x = _tmp; u0 = _tmp$1;
			if ((u0.high < u0x.high || (u0.high === u0x.high && u0.low < u0x.low))) {
				u1 = new Go$Uint64(u1.high + 0, u1.low + 1);
			}
			_tmp$2 = u0; _tmp$3 = (x$11 = new Go$Uint64(0, nsec), new Go$Uint64(u0.high + x$11.high, u0.low + x$11.low)); u0x = _tmp$2; u0 = _tmp$3;
			if ((u0.high < u0x.high || (u0.high === u0x.high && u0.low < u0x.low))) {
				u1 = new Go$Uint64(u1.high + 0, u1.low + 1);
			}
			d1$1 = new Go$Uint64(d.high, d.low);
			while (!((x$12 = go$shiftRightUint64(d1$1, 63), (x$12.high === 0 && x$12.low === 1)))) {
				d1$1 = go$shiftLeft64(d1$1, 1);
			}
			d0 = new Go$Uint64(0, 0);
			while (true) {
				qmod2 = 0;
				if ((u1.high > d1$1.high || (u1.high === d1$1.high && u1.low > d1$1.low)) || (u1.high === d1$1.high && u1.low === d1$1.low) && (u0.high > d0.high || (u0.high === d0.high && u0.low >= d0.low))) {
					qmod2 = 1;
					_tmp$4 = u0; _tmp$5 = new Go$Uint64(u0.high - d0.high, u0.low - d0.low); u0x = _tmp$4; u0 = _tmp$5;
					if ((u0.high > u0x.high || (u0.high === u0x.high && u0.low > u0x.low))) {
						u1 = new Go$Uint64(u1.high - 0, u1.low - 1);
					}
					u1 = (x$13 = d1$1, new Go$Uint64(u1.high - x$13.high, u1.low - x$13.low));
				}
				if ((d1$1.high === 0 && d1$1.low === 0) && (x$14 = new Go$Uint64(d.high, d.low), (d0.high === x$14.high && d0.low === x$14.low))) {
					break;
				}
				d0 = go$shiftRightUint64(d0, 1);
				d0 = (x$15 = go$shiftLeft64((new Go$Uint64(d1$1.high & 0, (d1$1.low & 1) >>> 0)), 63), new Go$Uint64(d0.high | x$15.high, (d0.low | x$15.low) >>> 0));
				d1$1 = go$shiftRightUint64(d1$1, 1);
			}
			r = new Duration(u0.high, u0.low);
		}
		if (neg && !((r.high === 0 && r.low === 0))) {
			qmod2 = (qmod2 ^ 1) >> 0;
			r = new Duration(d.high - r.high, d.low - r.low);
		}
		return [qmod2, r];
	};
	Location.Ptr.prototype.get = function() {
		var l;
		l = this;
		if (l === (go$ptrType(Location)).nil) {
			return utcLoc;
		}
		if (l === localLoc) {
			localOnce.Do(initLocal);
		}
		return l;
	};
	Location.prototype.get = function() { return this.go$val.get(); };
	Location.Ptr.prototype.String = function() {
		var l;
		l = this;
		return l.get().name;
	};
	Location.prototype.String = function() { return this.go$val.String(); };
	FixedZone = go$pkg.FixedZone = function(name, offset) {
		var l, x;
		l = new Location.Ptr(name, new (go$sliceType(zone))([new zone.Ptr(name, offset, false)]), new (go$sliceType(zoneTrans))([new zoneTrans.Ptr(new Go$Int64(-2147483648, 0), 0, false, false)]), new Go$Int64(-2147483648, 0), new Go$Int64(2147483647, 4294967295), (go$ptrType(zone)).nil);
		l.cacheZone = (x = l.zone, ((0 < 0 || 0 >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + 0]));
		return l;
	};
	Location.Ptr.prototype.lookup = function(sec) {
		var name, offset, isDST, start, end, l, zone$1, x, x$1, tx, lo, hi, _q, m, lim, x$2, x$3, zone$2;
		name = "";
		offset = 0;
		isDST = false;
		start = new Go$Int64(0, 0);
		end = new Go$Int64(0, 0);
		l = this;
		l = l.get();
		if (l.tx.length === 0) {
			name = "UTC";
			offset = 0;
			isDST = false;
			start = new Go$Int64(-2147483648, 0);
			end = new Go$Int64(2147483647, 4294967295);
			return [name, offset, isDST, start, end];
		}
		zone$1 = l.cacheZone;
		if (!(zone$1 === (go$ptrType(zone)).nil) && (x = l.cacheStart, (x.high < sec.high || (x.high === sec.high && x.low <= sec.low))) && (x$1 = l.cacheEnd, (sec.high < x$1.high || (sec.high === x$1.high && sec.low < x$1.low)))) {
			name = zone$1.name;
			offset = zone$1.offset;
			isDST = zone$1.isDST;
			start = l.cacheStart;
			end = l.cacheEnd;
			return [name, offset, isDST, start, end];
		}
		tx = l.tx;
		end = new Go$Int64(2147483647, 4294967295);
		lo = 0;
		hi = tx.length;
		while ((hi - lo >> 0) > 1) {
			m = lo + (_q = ((hi - lo >> 0)) / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0;
			lim = ((m < 0 || m >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + m]).when;
			if ((sec.high < lim.high || (sec.high === lim.high && sec.low < lim.low))) {
				end = lim;
				hi = m;
			} else {
				lo = m;
			}
		}
		zone$2 = (x$2 = l.zone, x$3 = ((lo < 0 || lo >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + lo]).index, ((x$3 < 0 || x$3 >= x$2.length) ? go$throwRuntimeError("index out of range") : x$2.array[x$2.offset + x$3]));
		name = zone$2.name;
		offset = zone$2.offset;
		isDST = zone$2.isDST;
		start = ((lo < 0 || lo >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + lo]).when;
		return [name, offset, isDST, start, end];
	};
	Location.prototype.lookup = function(sec) { return this.go$val.lookup(sec); };
	Location.Ptr.prototype.lookupName = function(name, unix) {
		var offset, isDST, ok, l, _ref, _i, i, x, zone$1, _tuple, x$1, nam, offset$1, isDST$1, _tmp, _tmp$1, _tmp$2, _ref$1, _i$1, i$1, x$2, zone$2, _tmp$3, _tmp$4, _tmp$5;
		offset = 0;
		isDST = false;
		ok = false;
		l = this;
		l = l.get();
		_ref = l.zone;
		_i = 0;
		while (_i < _ref.length) {
			i = _i;
			zone$1 = (x = l.zone, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
			if (zone$1.name === name) {
				_tuple = l.lookup((x$1 = new Go$Int64(0, zone$1.offset), new Go$Int64(unix.high - x$1.high, unix.low - x$1.low))); nam = _tuple[0]; offset$1 = _tuple[1]; isDST$1 = _tuple[2];
				if (nam === zone$1.name) {
					_tmp = offset$1; _tmp$1 = isDST$1; _tmp$2 = true; offset = _tmp; isDST = _tmp$1; ok = _tmp$2;
					return [offset, isDST, ok];
				}
			}
			_i++;
		}
		_ref$1 = l.zone;
		_i$1 = 0;
		while (_i$1 < _ref$1.length) {
			i$1 = _i$1;
			zone$2 = (x$2 = l.zone, ((i$1 < 0 || i$1 >= x$2.length) ? go$throwRuntimeError("index out of range") : x$2.array[x$2.offset + i$1]));
			if (zone$2.name === name) {
				_tmp$3 = zone$2.offset; _tmp$4 = zone$2.isDST; _tmp$5 = true; offset = _tmp$3; isDST = _tmp$4; ok = _tmp$5;
				return [offset, isDST, ok];
			}
			_i$1++;
		}
		return [offset, isDST, ok];
	};
	Location.prototype.lookupName = function(name, unix) { return this.go$val.lookupName(name, unix); };
	data.Ptr.prototype.read = function(n) {
		var d, p;
		d = this;
		if (d.p.length < n) {
			d.p = (go$sliceType(Go$Uint8)).nil;
			d.error = true;
			return (go$sliceType(Go$Uint8)).nil;
		}
		p = go$subslice(d.p, 0, n);
		d.p = go$subslice(d.p, n);
		return p;
	};
	data.prototype.read = function(n) { return this.go$val.read(n); };
	data.Ptr.prototype.big4 = function() {
		var n, ok, d, p, _tmp, _tmp$1, _tmp$2, _tmp$3;
		n = 0;
		ok = false;
		d = this;
		p = d.read(4);
		if (p.length < 4) {
			d.error = true;
			_tmp = 0; _tmp$1 = false; n = _tmp; ok = _tmp$1;
			return [n, ok];
		}
		_tmp$2 = (((((((((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]) >>> 0) << 24 >>> 0) | ((((1 < 0 || 1 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 1]) >>> 0) << 16 >>> 0)) >>> 0) | ((((2 < 0 || 2 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 2]) >>> 0) << 8 >>> 0)) >>> 0) | (((3 < 0 || 3 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 3]) >>> 0)) >>> 0; _tmp$3 = true; n = _tmp$2; ok = _tmp$3;
		return [n, ok];
	};
	data.prototype.big4 = function() { return this.go$val.big4(); };
	data.Ptr.prototype.byte$ = function() {
		var n, ok, d, p, _tmp, _tmp$1, _tmp$2, _tmp$3;
		n = 0;
		ok = false;
		d = this;
		p = d.read(1);
		if (p.length < 1) {
			d.error = true;
			_tmp = 0; _tmp$1 = false; n = _tmp; ok = _tmp$1;
			return [n, ok];
		}
		_tmp$2 = ((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]); _tmp$3 = true; n = _tmp$2; ok = _tmp$3;
		return [n, ok];
	};
	data.prototype.byte$ = function() { return this.go$val.byte$(); };
	byteString = function(p) {
		var i;
		i = 0;
		while (i < p.length) {
			if (((i < 0 || i >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + i]) === 0) {
				return go$bytesToString(go$subslice(p, 0, i));
			}
			i = i + 1 >> 0;
		}
		return go$bytesToString(p);
	};
	loadZoneData = function(bytes) {
		var l, err, d, magic, _tmp, _tmp$1, p, _tmp$2, _tmp$3, n, i, _tuple, nn, ok, _tmp$4, _tmp$5, x, txtimes, txzones, x$1, zonedata, abbrev, x$2, isstd, isutc, _tmp$6, _tmp$7, zone$1, _ref, _i, i$1, ok$1, n$1, _tuple$1, _tmp$8, _tmp$9, b, _tuple$2, _tmp$10, _tmp$11, _tuple$3, _tmp$12, _tmp$13, tx, _ref$1, _i$1, i$2, ok$2, n$2, _tuple$4, _tmp$14, _tmp$15, _tmp$16, _tmp$17, _tuple$5, sec, _ref$2, _i$2, i$3, x$3, x$4, x$5, x$6, x$7, x$8, _tmp$18, _tmp$19;
		l = (go$ptrType(Location)).nil;
		err = null;
		d = new data.Ptr(bytes, false);
		magic = d.read(4);
		if (!(go$bytesToString(magic) === "TZif")) {
			_tmp = (go$ptrType(Location)).nil; _tmp$1 = badData; l = _tmp; err = _tmp$1;
			return [l, err];
		}
		p = (go$sliceType(Go$Uint8)).nil;
		p = d.read(16);
		if (!((p.length === 16)) || !((((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]) === 0)) && !((((0 < 0 || 0 >= p.length) ? go$throwRuntimeError("index out of range") : p.array[p.offset + 0]) === 50))) {
			_tmp$2 = (go$ptrType(Location)).nil; _tmp$3 = badData; l = _tmp$2; err = _tmp$3;
			return [l, err];
		}
		n = go$makeNativeArray("Int", 6, function() { return 0; });
		i = 0;
		while (i < 6) {
			_tuple = d.big4(); nn = _tuple[0]; ok = _tuple[1];
			if (!ok) {
				_tmp$4 = (go$ptrType(Location)).nil; _tmp$5 = badData; l = _tmp$4; err = _tmp$5;
				return [l, err];
			}
			n[i] = (nn >> 0);
			i = i + 1 >> 0;
		}
		txtimes = new data.Ptr(d.read((x = n[3], (((x >>> 16 << 16) * 4 >> 0) + (x << 16 >>> 16) * 4) >> 0)), false);
		txzones = d.read(n[3]);
		zonedata = new data.Ptr(d.read((x$1 = n[4], (((x$1 >>> 16 << 16) * 6 >> 0) + (x$1 << 16 >>> 16) * 6) >> 0)), false);
		abbrev = d.read(n[5]);
		d.read((x$2 = n[2], (((x$2 >>> 16 << 16) * 8 >> 0) + (x$2 << 16 >>> 16) * 8) >> 0));
		isstd = d.read(n[1]);
		isutc = d.read(n[0]);
		if (d.error) {
			_tmp$6 = (go$ptrType(Location)).nil; _tmp$7 = badData; l = _tmp$6; err = _tmp$7;
			return [l, err];
		}
		zone$1 = (go$sliceType(zone)).make(n[4], 0, function() { return new zone.Ptr(); });
		_ref = zone$1;
		_i = 0;
		while (_i < _ref.length) {
			i$1 = _i;
			ok$1 = false;
			n$1 = 0;
			_tuple$1 = zonedata.big4(); n$1 = _tuple$1[0]; ok$1 = _tuple$1[1];
			if (!ok$1) {
				_tmp$8 = (go$ptrType(Location)).nil; _tmp$9 = badData; l = _tmp$8; err = _tmp$9;
				return [l, err];
			}
			((i$1 < 0 || i$1 >= zone$1.length) ? go$throwRuntimeError("index out of range") : zone$1.array[zone$1.offset + i$1]).offset = ((n$1 >> 0) >> 0);
			b = 0;
			_tuple$2 = zonedata.byte$(); b = _tuple$2[0]; ok$1 = _tuple$2[1];
			if (!ok$1) {
				_tmp$10 = (go$ptrType(Location)).nil; _tmp$11 = badData; l = _tmp$10; err = _tmp$11;
				return [l, err];
			}
			((i$1 < 0 || i$1 >= zone$1.length) ? go$throwRuntimeError("index out of range") : zone$1.array[zone$1.offset + i$1]).isDST = !((b === 0));
			_tuple$3 = zonedata.byte$(); b = _tuple$3[0]; ok$1 = _tuple$3[1];
			if (!ok$1 || (b >> 0) >= abbrev.length) {
				_tmp$12 = (go$ptrType(Location)).nil; _tmp$13 = badData; l = _tmp$12; err = _tmp$13;
				return [l, err];
			}
			((i$1 < 0 || i$1 >= zone$1.length) ? go$throwRuntimeError("index out of range") : zone$1.array[zone$1.offset + i$1]).name = byteString(go$subslice(abbrev, b));
			_i++;
		}
		tx = (go$sliceType(zoneTrans)).make(n[3], 0, function() { return new zoneTrans.Ptr(); });
		_ref$1 = tx;
		_i$1 = 0;
		while (_i$1 < _ref$1.length) {
			i$2 = _i$1;
			ok$2 = false;
			n$2 = 0;
			_tuple$4 = txtimes.big4(); n$2 = _tuple$4[0]; ok$2 = _tuple$4[1];
			if (!ok$2) {
				_tmp$14 = (go$ptrType(Location)).nil; _tmp$15 = badData; l = _tmp$14; err = _tmp$15;
				return [l, err];
			}
			((i$2 < 0 || i$2 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$2]).when = new Go$Int64(0, (n$2 >> 0));
			if ((((i$2 < 0 || i$2 >= txzones.length) ? go$throwRuntimeError("index out of range") : txzones.array[txzones.offset + i$2]) >> 0) >= zone$1.length) {
				_tmp$16 = (go$ptrType(Location)).nil; _tmp$17 = badData; l = _tmp$16; err = _tmp$17;
				return [l, err];
			}
			((i$2 < 0 || i$2 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$2]).index = ((i$2 < 0 || i$2 >= txzones.length) ? go$throwRuntimeError("index out of range") : txzones.array[txzones.offset + i$2]);
			if (i$2 < isstd.length) {
				((i$2 < 0 || i$2 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$2]).isstd = !((((i$2 < 0 || i$2 >= isstd.length) ? go$throwRuntimeError("index out of range") : isstd.array[isstd.offset + i$2]) === 0));
			}
			if (i$2 < isutc.length) {
				((i$2 < 0 || i$2 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$2]).isutc = !((((i$2 < 0 || i$2 >= isutc.length) ? go$throwRuntimeError("index out of range") : isutc.array[isutc.offset + i$2]) === 0));
			}
			_i$1++;
		}
		if (tx.length === 0) {
			tx = go$append(tx, new zoneTrans.Ptr(new Go$Int64(-2147483648, 0), 0, false, false));
		}
		l = new Location.Ptr("", zone$1, tx, new Go$Int64(0, 0), new Go$Int64(0, 0), (go$ptrType(zone)).nil);
		_tuple$5 = now(); sec = _tuple$5[0];
		_ref$2 = tx;
		_i$2 = 0;
		while (_i$2 < _ref$2.length) {
			i$3 = _i$2;
			if ((x$3 = ((i$3 < 0 || i$3 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$3]).when, (x$3.high < sec.high || (x$3.high === sec.high && x$3.low <= sec.low))) && (((i$3 + 1 >> 0) === tx.length) || (x$4 = (x$5 = i$3 + 1 >> 0, ((x$5 < 0 || x$5 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + x$5])).when, (sec.high < x$4.high || (sec.high === x$4.high && sec.low < x$4.low))))) {
				l.cacheStart = ((i$3 < 0 || i$3 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$3]).when;
				l.cacheEnd = new Go$Int64(2147483647, 4294967295);
				if ((i$3 + 1 >> 0) < tx.length) {
					l.cacheEnd = (x$6 = i$3 + 1 >> 0, ((x$6 < 0 || x$6 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + x$6])).when;
				}
				l.cacheZone = (x$7 = l.zone, x$8 = ((i$3 < 0 || i$3 >= tx.length) ? go$throwRuntimeError("index out of range") : tx.array[tx.offset + i$3]).index, ((x$8 < 0 || x$8 >= x$7.length) ? go$throwRuntimeError("index out of range") : x$7.array[x$7.offset + x$8]));
			}
			_i$2++;
		}
		_tmp$18 = l; _tmp$19 = null; l = _tmp$18; err = _tmp$19;
		return [l, err];
	};
	loadZoneFile = function(dir, name) {
		var l, err, _tuple, _tuple$1, buf, _tuple$2;
		l = (go$ptrType(Location)).nil;
		err = null;
		if (dir.length > 4 && dir.substring((dir.length - 4 >> 0)) === ".zip") {
			_tuple = loadZoneZip(dir, name); l = _tuple[0]; err = _tuple[1];
			return [l, err];
		}
		if (!(dir === "")) {
			name = dir + "/" + name;
		}
		_tuple$1 = readFile(name); buf = _tuple$1[0]; err = _tuple$1[1];
		if (!(go$interfaceIsEqual(err, null))) {
			return [l, err];
		}
		_tuple$2 = loadZoneData(buf); l = _tuple$2[0]; err = _tuple$2[1];
		return [l, err];
	};
	get4 = function(b) {
		if (b.length < 4) {
			return 0;
		}
		return (((((0 < 0 || 0 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + 0]) >> 0) | ((((1 < 0 || 1 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + 1]) >> 0) << 8 >> 0)) | ((((2 < 0 || 2 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + 2]) >> 0) << 16 >> 0)) | ((((3 < 0 || 3 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + 3]) >> 0) << 24 >> 0);
	};
	get2 = function(b) {
		if (b.length < 2) {
			return 0;
		}
		return (((0 < 0 || 0 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + 0]) >> 0) | ((((1 < 0 || 1 >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + 1]) >> 0) << 8 >> 0);
	};
	loadZoneZip = function(zipfile, name) {
		var l, err, _tuple, fd, _tmp, _tmp$1, buf, err$1, _tmp$2, _tmp$3, n, size, off, err$2, _tmp$4, _tmp$5, i, meth, size$1, namelen, xlen, fclen, off$1, zname, _tmp$6, _tmp$7, err$3, _tmp$8, _tmp$9, err$4, _tmp$10, _tmp$11, _tuple$1, _tmp$12, _tmp$13;
		l = (go$ptrType(Location)).nil;
		err = null;
		var go$deferred = [];
		try {
			_tuple = open(zipfile); fd = _tuple[0]; err = _tuple[1];
			if (!(go$interfaceIsEqual(err, null))) {
				_tmp = (go$ptrType(Location)).nil; _tmp$1 = errors.New("open " + zipfile + ": " + err.Error()); l = _tmp; err = _tmp$1;
				return [l, err];
			}
			go$deferred.push({ fun: closefd, args: [fd] });
			buf = (go$sliceType(Go$Uint8)).make(22, 0, function() { return 0; });
			err$1 = preadn(fd, buf, -22);
			if (!(go$interfaceIsEqual(err$1, null)) || !((get4(buf) === 101010256))) {
				_tmp$2 = (go$ptrType(Location)).nil; _tmp$3 = errors.New("corrupt zip file " + zipfile); l = _tmp$2; err = _tmp$3;
				return [l, err];
			}
			n = get2(go$subslice(buf, 10));
			size = get4(go$subslice(buf, 12));
			off = get4(go$subslice(buf, 16));
			buf = (go$sliceType(Go$Uint8)).make(size, 0, function() { return 0; });
			err$2 = preadn(fd, buf, off);
			if (!(go$interfaceIsEqual(err$2, null))) {
				_tmp$4 = (go$ptrType(Location)).nil; _tmp$5 = errors.New("corrupt zip file " + zipfile); l = _tmp$4; err = _tmp$5;
				return [l, err];
			}
			i = 0;
			while (i < n) {
				if (!((get4(buf) === 33639248))) {
					break;
				}
				meth = get2(go$subslice(buf, 10));
				size$1 = get4(go$subslice(buf, 24));
				namelen = get2(go$subslice(buf, 28));
				xlen = get2(go$subslice(buf, 30));
				fclen = get2(go$subslice(buf, 32));
				off$1 = get4(go$subslice(buf, 42));
				zname = go$subslice(buf, 46, (46 + namelen >> 0));
				buf = go$subslice(buf, (((46 + namelen >> 0) + xlen >> 0) + fclen >> 0));
				if (!(go$bytesToString(zname) === name)) {
					i = i + 1 >> 0;
					continue;
				}
				if (!((meth === 0))) {
					_tmp$6 = (go$ptrType(Location)).nil; _tmp$7 = errors.New("unsupported compression for " + name + " in " + zipfile); l = _tmp$6; err = _tmp$7;
					return [l, err];
				}
				buf = (go$sliceType(Go$Uint8)).make((30 + namelen >> 0), 0, function() { return 0; });
				err$3 = preadn(fd, buf, off$1);
				if (!(go$interfaceIsEqual(err$3, null)) || !((get4(buf) === 67324752)) || !((get2(go$subslice(buf, 8)) === meth)) || !((get2(go$subslice(buf, 26)) === namelen)) || !(go$bytesToString(go$subslice(buf, 30, (30 + namelen >> 0))) === name)) {
					_tmp$8 = (go$ptrType(Location)).nil; _tmp$9 = errors.New("corrupt zip file " + zipfile); l = _tmp$8; err = _tmp$9;
					return [l, err];
				}
				xlen = get2(go$subslice(buf, 28));
				buf = (go$sliceType(Go$Uint8)).make(size$1, 0, function() { return 0; });
				err$4 = preadn(fd, buf, ((off$1 + 30 >> 0) + namelen >> 0) + xlen >> 0);
				if (!(go$interfaceIsEqual(err$4, null))) {
					_tmp$10 = (go$ptrType(Location)).nil; _tmp$11 = errors.New("corrupt zip file " + zipfile); l = _tmp$10; err = _tmp$11;
					return [l, err];
				}
				_tuple$1 = loadZoneData(buf); l = _tuple$1[0]; err = _tuple$1[1];
				return [l, err];
			}
			_tmp$12 = (go$ptrType(Location)).nil; _tmp$13 = errors.New("cannot find " + name + " in zip file " + zipfile); l = _tmp$12; err = _tmp$13;
			return [l, err];
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
			return [l, err];
		}
	};
	initLocal = function() {
		var _tuple, tz, ok, _tuple$1, z, err, _struct, _tuple$2, z$1, err$1, _struct$1;
		_tuple = syscall.Getenv("TZ"); tz = _tuple[0]; ok = _tuple[1];
		if (!ok) {
			_tuple$1 = loadZoneFile("", "/etc/localtime"); z = _tuple$1[0]; err = _tuple$1[1];
			if (go$interfaceIsEqual(err, null)) {
				localLoc = (_struct = z, new Location.Ptr(_struct.name, _struct.zone, _struct.tx, _struct.cacheStart, _struct.cacheEnd, _struct.cacheZone));
				localLoc.name = "Local";
				return;
			}
		} else if (!(tz === "") && !(tz === "UTC")) {
			_tuple$2 = loadLocation(tz); z$1 = _tuple$2[0]; err$1 = _tuple$2[1];
			if (go$interfaceIsEqual(err$1, null)) {
				localLoc = (_struct$1 = z$1, new Location.Ptr(_struct$1.name, _struct$1.zone, _struct$1.tx, _struct$1.cacheStart, _struct$1.cacheEnd, _struct$1.cacheZone));
				return;
			}
		}
		localLoc.name = "UTC";
	};
	loadLocation = function(name) {
		var _ref, _i, zoneDir, _tuple, z, err;
		_ref = zoneDirs;
		_i = 0;
		while (_i < _ref.length) {
			zoneDir = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			_tuple = loadZoneFile(zoneDir, name); z = _tuple[0]; err = _tuple[1];
			if (go$interfaceIsEqual(err, null)) {
				z.name = name;
				return [z, null];
			}
			_i++;
		}
		return [(go$ptrType(Location)).nil, errors.New("unknown time zone " + name)];
	};
	go$pkg.init = function() {
		(go$ptrType(ParseError)).methods = [["Error", "", [], [Go$String], false, -1]];
		ParseError.init([["Layout", "Layout", "", Go$String, ""], ["Value", "Value", "", Go$String, ""], ["LayoutElem", "LayoutElem", "", Go$String, ""], ["ValueElem", "ValueElem", "", Go$String, ""], ["Message", "Message", "", Go$String, ""]]);
		Time.methods = [["Add", "", [Duration], [Time], false, -1], ["AddDate", "", [Go$Int, Go$Int, Go$Int], [Time], false, -1], ["After", "", [Time], [Go$Bool], false, -1], ["Before", "", [Time], [Go$Bool], false, -1], ["Clock", "", [], [Go$Int, Go$Int, Go$Int], false, -1], ["Date", "", [], [Go$Int, Month, Go$Int], false, -1], ["Day", "", [], [Go$Int], false, -1], ["Equal", "", [Time], [Go$Bool], false, -1], ["Format", "", [Go$String], [Go$String], false, -1], ["GobEncode", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["Hour", "", [], [Go$Int], false, -1], ["ISOWeek", "", [], [Go$Int, Go$Int], false, -1], ["In", "", [(go$ptrType(Location))], [Time], false, -1], ["IsZero", "", [], [Go$Bool], false, -1], ["Local", "", [], [Time], false, -1], ["Location", "", [], [(go$ptrType(Location))], false, -1], ["MarshalBinary", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["MarshalJSON", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["MarshalText", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["Minute", "", [], [Go$Int], false, -1], ["Month", "", [], [Month], false, -1], ["Nanosecond", "", [], [Go$Int], false, -1], ["Round", "", [Duration], [Time], false, -1], ["Second", "", [], [Go$Int], false, -1], ["String", "", [], [Go$String], false, -1], ["Sub", "", [Time], [Duration], false, -1], ["Truncate", "", [Duration], [Time], false, -1], ["UTC", "", [], [Time], false, -1], ["Unix", "", [], [Go$Int64], false, -1], ["UnixNano", "", [], [Go$Int64], false, -1], ["Weekday", "", [], [Weekday], false, -1], ["Year", "", [], [Go$Int], false, -1], ["YearDay", "", [], [Go$Int], false, -1], ["Zone", "", [], [Go$String, Go$Int], false, -1], ["abs", "time", [], [Go$Uint64], false, -1], ["date", "time", [Go$Bool], [Go$Int, Month, Go$Int, Go$Int], false, -1], ["locabs", "time", [], [Go$String, Go$Int, Go$Uint64], false, -1]];
		(go$ptrType(Time)).methods = [["Add", "", [Duration], [Time], false, -1], ["AddDate", "", [Go$Int, Go$Int, Go$Int], [Time], false, -1], ["After", "", [Time], [Go$Bool], false, -1], ["Before", "", [Time], [Go$Bool], false, -1], ["Clock", "", [], [Go$Int, Go$Int, Go$Int], false, -1], ["Date", "", [], [Go$Int, Month, Go$Int], false, -1], ["Day", "", [], [Go$Int], false, -1], ["Equal", "", [Time], [Go$Bool], false, -1], ["Format", "", [Go$String], [Go$String], false, -1], ["GobDecode", "", [(go$sliceType(Go$Uint8))], [go$error], false, -1], ["GobEncode", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["Hour", "", [], [Go$Int], false, -1], ["ISOWeek", "", [], [Go$Int, Go$Int], false, -1], ["In", "", [(go$ptrType(Location))], [Time], false, -1], ["IsZero", "", [], [Go$Bool], false, -1], ["Local", "", [], [Time], false, -1], ["Location", "", [], [(go$ptrType(Location))], false, -1], ["MarshalBinary", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["MarshalJSON", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["MarshalText", "", [], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["Minute", "", [], [Go$Int], false, -1], ["Month", "", [], [Month], false, -1], ["Nanosecond", "", [], [Go$Int], false, -1], ["Round", "", [Duration], [Time], false, -1], ["Second", "", [], [Go$Int], false, -1], ["String", "", [], [Go$String], false, -1], ["Sub", "", [Time], [Duration], false, -1], ["Truncate", "", [Duration], [Time], false, -1], ["UTC", "", [], [Time], false, -1], ["Unix", "", [], [Go$Int64], false, -1], ["UnixNano", "", [], [Go$Int64], false, -1], ["UnmarshalBinary", "", [(go$sliceType(Go$Uint8))], [go$error], false, -1], ["UnmarshalJSON", "", [(go$sliceType(Go$Uint8))], [go$error], false, -1], ["UnmarshalText", "", [(go$sliceType(Go$Uint8))], [go$error], false, -1], ["Weekday", "", [], [Weekday], false, -1], ["Year", "", [], [Go$Int], false, -1], ["YearDay", "", [], [Go$Int], false, -1], ["Zone", "", [], [Go$String, Go$Int], false, -1], ["abs", "time", [], [Go$Uint64], false, -1], ["date", "time", [Go$Bool], [Go$Int, Month, Go$Int, Go$Int], false, -1], ["locabs", "time", [], [Go$String, Go$Int, Go$Uint64], false, -1]];
		Time.init([["sec", "sec", "time", Go$Int64, ""], ["nsec", "nsec", "time", Go$Uintptr, ""], ["loc", "loc", "time", (go$ptrType(Location)), ""]]);
		Month.methods = [["String", "", [], [Go$String], false, -1]];
		(go$ptrType(Month)).methods = [["String", "", [], [Go$String], false, -1]];
		Weekday.methods = [["String", "", [], [Go$String], false, -1]];
		(go$ptrType(Weekday)).methods = [["String", "", [], [Go$String], false, -1]];
		Duration.methods = [["Hours", "", [], [Go$Float64], false, -1], ["Minutes", "", [], [Go$Float64], false, -1], ["Nanoseconds", "", [], [Go$Int64], false, -1], ["Seconds", "", [], [Go$Float64], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(Duration)).methods = [["Hours", "", [], [Go$Float64], false, -1], ["Minutes", "", [], [Go$Float64], false, -1], ["Nanoseconds", "", [], [Go$Int64], false, -1], ["Seconds", "", [], [Go$Float64], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(Location)).methods = [["String", "", [], [Go$String], false, -1], ["get", "time", [], [(go$ptrType(Location))], false, -1], ["lookup", "time", [Go$Int64], [Go$String, Go$Int, Go$Bool, Go$Int64, Go$Int64], false, -1], ["lookupName", "time", [Go$String, Go$Int64], [Go$Int, Go$Bool, Go$Bool], false, -1]];
		Location.init([["name", "name", "time", Go$String, ""], ["zone", "zone", "time", (go$sliceType(zone)), ""], ["tx", "tx", "time", (go$sliceType(zoneTrans)), ""], ["cacheStart", "cacheStart", "time", Go$Int64, ""], ["cacheEnd", "cacheEnd", "time", Go$Int64, ""], ["cacheZone", "cacheZone", "time", (go$ptrType(zone)), ""]]);
		zone.init([["name", "name", "time", Go$String, ""], ["offset", "offset", "time", Go$Int, ""], ["isDST", "isDST", "time", Go$Bool, ""]]);
		zoneTrans.init([["when", "when", "time", Go$Int64, ""], ["index", "index", "time", Go$Uint8, ""], ["isstd", "isstd", "time", Go$Bool, ""], ["isutc", "isutc", "time", Go$Bool, ""]]);
		(go$ptrType(data)).methods = [["big4", "time", [], [Go$Uint32, Go$Bool], false, -1], ["byte", "time", [], [Go$Uint8, Go$Bool], false, -1], ["read", "time", [Go$Int], [(go$sliceType(Go$Uint8))], false, -1]];
		data.init([["p", "p", "time", (go$sliceType(Go$Uint8)), ""], ["error", "error", "time", Go$Bool, ""]]);
		localLoc = new Location.Ptr();
		localOnce = new sync.Once.Ptr();
		std0x = go$toNativeArray("Int", [260, 265, 524, 526, 528, 274]);
		longDayNames = new (go$sliceType(Go$String))(["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]);
		shortDayNames = new (go$sliceType(Go$String))(["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]);
		shortMonthNames = new (go$sliceType(Go$String))(["---", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]);
		longMonthNames = new (go$sliceType(Go$String))(["---", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"]);
		atoiError = errors.New("time: invalid number");
		errBad = errors.New("bad value for field");
		errLeadingInt = errors.New("time: bad [0-9]*");
		months = go$toNativeArray("String", ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"]);
		days = go$toNativeArray("String", ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]);
		daysBefore = go$toNativeArray("Int32", [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365]);
		utcLoc = new Location.Ptr("UTC", (go$sliceType(zone)).nil, (go$sliceType(zoneTrans)).nil, new Go$Int64(0, 0), new Go$Int64(0, 0), (go$ptrType(zone)).nil);
		go$pkg.UTC = utcLoc;
		go$pkg.Local = localLoc;
		var _tuple;
		_tuple = syscall.Getenv("ZONEINFO"); zoneinfo = _tuple[0];
		badData = errors.New("malformed time zone information");
		zoneDirs = new (go$sliceType(Go$String))(["/usr/share/zoneinfo/", "/usr/share/lib/zoneinfo/", "/usr/lib/locale/TZ/", runtime.GOROOT() + "/lib/time/zoneinfo.zip"]);
	}
	return go$pkg;
})();
go$packages["os"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], io = go$packages["io"], syscall = go$packages["syscall"], time = go$packages["time"], errors = go$packages["errors"], runtime = go$packages["runtime"], atomic = go$packages["sync/atomic"], sync = go$packages["sync"], PathError, SyscallError, File, file, dirInfo, FileInfo, FileMode, fileStat, NewSyscallError, sigpipe, syscallMode, NewFile, epipecheck, Lstat, basename, useSyscallwdDarwin, fileInfoFromStat, timespecToTime, lstat, useSyscallwd;
	PathError = go$pkg.PathError = go$newType(0, "Struct", "os.PathError", "PathError", "os", function(Op_, Path_, Err_) {
		this.go$val = this;
		this.Op = Op_ !== undefined ? Op_ : "";
		this.Path = Path_ !== undefined ? Path_ : "";
		this.Err = Err_ !== undefined ? Err_ : null;
	});
	SyscallError = go$pkg.SyscallError = go$newType(0, "Struct", "os.SyscallError", "SyscallError", "os", function(Syscall_, Err_) {
		this.go$val = this;
		this.Syscall = Syscall_ !== undefined ? Syscall_ : "";
		this.Err = Err_ !== undefined ? Err_ : null;
	});
	File = go$pkg.File = go$newType(0, "Struct", "os.File", "File", "os", function(file_) {
		this.go$val = this;
		this.file = file_ !== undefined ? file_ : (go$ptrType(file)).nil;
	});
	file = go$pkg.file = go$newType(0, "Struct", "os.file", "file", "os", function(fd_, name_, dirinfo_, nepipe_) {
		this.go$val = this;
		this.fd = fd_ !== undefined ? fd_ : 0;
		this.name = name_ !== undefined ? name_ : "";
		this.dirinfo = dirinfo_ !== undefined ? dirinfo_ : (go$ptrType(dirInfo)).nil;
		this.nepipe = nepipe_ !== undefined ? nepipe_ : 0;
	});
	dirInfo = go$pkg.dirInfo = go$newType(0, "Struct", "os.dirInfo", "dirInfo", "os", function(buf_, nbuf_, bufp_) {
		this.go$val = this;
		this.buf = buf_ !== undefined ? buf_ : (go$sliceType(Go$Uint8)).nil;
		this.nbuf = nbuf_ !== undefined ? nbuf_ : 0;
		this.bufp = bufp_ !== undefined ? bufp_ : 0;
	});
	FileInfo = go$pkg.FileInfo = go$newType(8, "Interface", "os.FileInfo", "FileInfo", "os", null);
	FileMode = go$pkg.FileMode = go$newType(4, "Uint32", "os.FileMode", "FileMode", "os", null);
	fileStat = go$pkg.fileStat = go$newType(0, "Struct", "os.fileStat", "fileStat", "os", function(name_, size_, mode_, modTime_, sys_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : "";
		this.size = size_ !== undefined ? size_ : new Go$Int64(0, 0);
		this.mode = mode_ !== undefined ? mode_ : 0;
		this.modTime = modTime_ !== undefined ? modTime_ : new time.Time.Ptr();
		this.sys = sys_ !== undefined ? sys_ : null;
	});
	File.Ptr.prototype.readdirnames = function(n) {
		var names, err, f, d, size, errno, _tuple, _tmp, _tmp$1, _tmp$2, _tmp$3, nb, nc, _tuple$1, _tmp$4, _tmp$5, _tmp$6, _tmp$7;
		names = (go$sliceType(Go$String)).nil;
		err = null;
		f = this;
		if (f.file.dirinfo === (go$ptrType(dirInfo)).nil) {
			f.file.dirinfo = new dirInfo.Ptr();
			f.file.dirinfo.buf = (go$sliceType(Go$Uint8)).make(4096, 0, function() { return 0; });
		}
		d = f.file.dirinfo;
		size = n;
		if (size <= 0) {
			size = 100;
			n = -1;
		}
		names = (go$sliceType(Go$String)).make(0, size, function() { return ""; });
		while (!((n === 0))) {
			if (d.bufp >= d.nbuf) {
				d.bufp = 0;
				errno = null;
				_tuple = syscall.ReadDirent(f.file.fd, d.buf); d.nbuf = _tuple[0]; errno = _tuple[1];
				if (!(go$interfaceIsEqual(errno, null))) {
					_tmp = names; _tmp$1 = NewSyscallError("readdirent", errno); names = _tmp; err = _tmp$1;
					return [names, err];
				}
				if (d.nbuf <= 0) {
					break;
				}
			}
			_tmp$2 = 0; _tmp$3 = 0; nb = _tmp$2; nc = _tmp$3;
			_tuple$1 = syscall.ParseDirent(go$subslice(d.buf, d.bufp, d.nbuf), n, names); nb = _tuple$1[0]; nc = _tuple$1[1]; names = _tuple$1[2];
			d.bufp = d.bufp + (nb) >> 0;
			n = n - (nc) >> 0;
		}
		if (n >= 0 && (names.length === 0)) {
			_tmp$4 = names; _tmp$5 = io.EOF; names = _tmp$4; err = _tmp$5;
			return [names, err];
		}
		_tmp$6 = names; _tmp$7 = null; names = _tmp$6; err = _tmp$7;
		return [names, err];
	};
	File.prototype.readdirnames = function(n) { return this.go$val.readdirnames(n); };
	File.Ptr.prototype.Readdir = function(n) {
		var fi, err, f, _tmp, _tmp$1, _tuple;
		fi = (go$sliceType(FileInfo)).nil;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = (go$sliceType(FileInfo)).nil; _tmp$1 = go$pkg.ErrInvalid; fi = _tmp; err = _tmp$1;
			return [fi, err];
		}
		_tuple = f.readdir(n); fi = _tuple[0]; err = _tuple[1];
		return [fi, err];
	};
	File.prototype.Readdir = function(n) { return this.go$val.Readdir(n); };
	File.Ptr.prototype.Readdirnames = function(n) {
		var names, err, f, _tmp, _tmp$1, _tuple;
		names = (go$sliceType(Go$String)).nil;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = (go$sliceType(Go$String)).nil; _tmp$1 = go$pkg.ErrInvalid; names = _tmp; err = _tmp$1;
			return [names, err];
		}
		_tuple = f.readdirnames(n); names = _tuple[0]; err = _tuple[1];
		return [names, err];
	};
	File.prototype.Readdirnames = function(n) { return this.go$val.Readdirnames(n); };
	PathError.Ptr.prototype.Error = function() {
		var e;
		e = this;
		return e.Op + " " + e.Path + ": " + e.Err.Error();
	};
	PathError.prototype.Error = function() { return this.go$val.Error(); };
	SyscallError.Ptr.prototype.Error = function() {
		var e;
		e = this;
		return e.Syscall + ": " + e.Err.Error();
	};
	SyscallError.prototype.Error = function() { return this.go$val.Error(); };
	NewSyscallError = go$pkg.NewSyscallError = function(syscall$1, err) {
		if (go$interfaceIsEqual(err, null)) {
			return null;
		}
		return new SyscallError.Ptr(syscall$1, err);
	};
	File.Ptr.prototype.Name = function() {
		var f;
		f = this;
		return f.file.name;
	};
	File.prototype.Name = function() { return this.go$val.Name(); };
	File.Ptr.prototype.Read = function(b) {
		var n, err, f, _tmp, _tmp$1, _tuple, e, _tmp$2, _tmp$3, _tmp$4, _tmp$5;
		n = 0;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = 0; _tmp$1 = go$pkg.ErrInvalid; n = _tmp; err = _tmp$1;
			return [n, err];
		}
		_tuple = f.read(b); n = _tuple[0]; e = _tuple[1];
		if (n < 0) {
			n = 0;
		}
		if ((n === 0) && b.length > 0 && go$interfaceIsEqual(e, null)) {
			_tmp$2 = 0; _tmp$3 = io.EOF; n = _tmp$2; err = _tmp$3;
			return [n, err];
		}
		if (!(go$interfaceIsEqual(e, null))) {
			err = new PathError.Ptr("read", f.file.name, e);
		}
		_tmp$4 = n; _tmp$5 = err; n = _tmp$4; err = _tmp$5;
		return [n, err];
	};
	File.prototype.Read = function(b) { return this.go$val.Read(b); };
	File.Ptr.prototype.ReadAt = function(b, off) {
		var n, err, f, _tmp, _tmp$1, _tuple, m, e, _tmp$2, _tmp$3, x;
		n = 0;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = 0; _tmp$1 = go$pkg.ErrInvalid; n = _tmp; err = _tmp$1;
			return [n, err];
		}
		while (b.length > 0) {
			_tuple = f.pread(b, off); m = _tuple[0]; e = _tuple[1];
			if ((m === 0) && go$interfaceIsEqual(e, null)) {
				_tmp$2 = n; _tmp$3 = io.EOF; n = _tmp$2; err = _tmp$3;
				return [n, err];
			}
			if (!(go$interfaceIsEqual(e, null))) {
				err = new PathError.Ptr("read", f.file.name, e);
				break;
			}
			n = n + (m) >> 0;
			b = go$subslice(b, m);
			off = (x = new Go$Int64(0, m), new Go$Int64(off.high + x.high, off.low + x.low));
		}
		return [n, err];
	};
	File.prototype.ReadAt = function(b, off) { return this.go$val.ReadAt(b, off); };
	File.Ptr.prototype.Write = function(b) {
		var n, err, f, _tmp, _tmp$1, _tuple, e, _tmp$2, _tmp$3;
		n = 0;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = 0; _tmp$1 = go$pkg.ErrInvalid; n = _tmp; err = _tmp$1;
			return [n, err];
		}
		_tuple = f.write(b); n = _tuple[0]; e = _tuple[1];
		if (n < 0) {
			n = 0;
		}
		epipecheck(f, e);
		if (!(go$interfaceIsEqual(e, null))) {
			err = new PathError.Ptr("write", f.file.name, e);
		}
		_tmp$2 = n; _tmp$3 = err; n = _tmp$2; err = _tmp$3;
		return [n, err];
	};
	File.prototype.Write = function(b) { return this.go$val.Write(b); };
	File.Ptr.prototype.WriteAt = function(b, off) {
		var n, err, f, _tmp, _tmp$1, _tuple, m, e, x;
		n = 0;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = 0; _tmp$1 = go$pkg.ErrInvalid; n = _tmp; err = _tmp$1;
			return [n, err];
		}
		while (b.length > 0) {
			_tuple = f.pwrite(b, off); m = _tuple[0]; e = _tuple[1];
			if (!(go$interfaceIsEqual(e, null))) {
				err = new PathError.Ptr("write", f.file.name, e);
				break;
			}
			n = n + (m) >> 0;
			b = go$subslice(b, m);
			off = (x = new Go$Int64(0, m), new Go$Int64(off.high + x.high, off.low + x.low));
		}
		return [n, err];
	};
	File.prototype.WriteAt = function(b, off) { return this.go$val.WriteAt(b, off); };
	File.Ptr.prototype.Seek = function(offset, whence) {
		var ret, err, f, _tmp, _tmp$1, _tuple, r, e, _tmp$2, _tmp$3, _tmp$4, _tmp$5;
		ret = new Go$Int64(0, 0);
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = new Go$Int64(0, 0); _tmp$1 = go$pkg.ErrInvalid; ret = _tmp; err = _tmp$1;
			return [ret, err];
		}
		_tuple = f.seek(offset, whence); r = _tuple[0]; e = _tuple[1];
		if (go$interfaceIsEqual(e, null) && !(f.file.dirinfo === (go$ptrType(dirInfo)).nil) && !((r.high === 0 && r.low === 0))) {
			e = new syscall.Errno(21);
		}
		if (!(go$interfaceIsEqual(e, null))) {
			_tmp$2 = new Go$Int64(0, 0); _tmp$3 = new PathError.Ptr("seek", f.file.name, e); ret = _tmp$2; err = _tmp$3;
			return [ret, err];
		}
		_tmp$4 = r; _tmp$5 = null; ret = _tmp$4; err = _tmp$5;
		return [ret, err];
	};
	File.prototype.Seek = function(offset, whence) { return this.go$val.Seek(offset, whence); };
	File.Ptr.prototype.WriteString = function(s) {
		var ret, err, f, _tmp, _tmp$1, _tuple;
		ret = 0;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = 0; _tmp$1 = go$pkg.ErrInvalid; ret = _tmp; err = _tmp$1;
			return [ret, err];
		}
		_tuple = f.Write(new (go$sliceType(Go$Uint8))(go$stringToBytes(s))); ret = _tuple[0]; err = _tuple[1];
		return [ret, err];
	};
	File.prototype.WriteString = function(s) { return this.go$val.WriteString(s); };
	File.Ptr.prototype.Chdir = function() {
		var f, e;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			return go$pkg.ErrInvalid;
		}
		e = syscall.Fchdir(f.file.fd);
		if (!(go$interfaceIsEqual(e, null))) {
			return new PathError.Ptr("chdir", f.file.name, e);
		}
		return null;
	};
	File.prototype.Chdir = function() { return this.go$val.Chdir(); };
	sigpipe = function() {
		throw go$panic("Native function not implemented: sigpipe");
	};
	syscallMode = function(i) {
		var o;
		o = 0;
		o = (o | (((new FileMode(i)).Perm() >>> 0))) >>> 0;
		if (!((((i & 8388608) >>> 0) === 0))) {
			o = (o | 2048) >>> 0;
		}
		if (!((((i & 4194304) >>> 0) === 0))) {
			o = (o | 1024) >>> 0;
		}
		if (!((((i & 1048576) >>> 0) === 0))) {
			o = (o | 512) >>> 0;
		}
		return o;
	};
	File.Ptr.prototype.Chmod = function(mode) {
		var f, e;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			return go$pkg.ErrInvalid;
		}
		e = syscall.Fchmod(f.file.fd, syscallMode(mode));
		if (!(go$interfaceIsEqual(e, null))) {
			return new PathError.Ptr("chmod", f.file.name, e);
		}
		return null;
	};
	File.prototype.Chmod = function(mode) { return this.go$val.Chmod(mode); };
	File.Ptr.prototype.Chown = function(uid, gid) {
		var f, e;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			return go$pkg.ErrInvalid;
		}
		e = syscall.Fchown(f.file.fd, uid, gid);
		if (!(go$interfaceIsEqual(e, null))) {
			return new PathError.Ptr("chown", f.file.name, e);
		}
		return null;
	};
	File.prototype.Chown = function(uid, gid) { return this.go$val.Chown(uid, gid); };
	File.Ptr.prototype.Truncate = function(size) {
		var f, e;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			return go$pkg.ErrInvalid;
		}
		e = syscall.Ftruncate(f.file.fd, size);
		if (!(go$interfaceIsEqual(e, null))) {
			return new PathError.Ptr("truncate", f.file.name, e);
		}
		return null;
	};
	File.prototype.Truncate = function(size) { return this.go$val.Truncate(size); };
	File.Ptr.prototype.Sync = function() {
		var err, f, e;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			err = new syscall.Errno(22);
			return err;
		}
		e = syscall.Fsync(f.file.fd);
		if (!(go$interfaceIsEqual(e, null))) {
			err = NewSyscallError("fsync", e);
			return err;
		}
		err = null;
		return err;
	};
	File.prototype.Sync = function() { return this.go$val.Sync(); };
	File.Ptr.prototype.Fd = function() {
		var f;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			return 4294967295;
		}
		return (f.file.fd >>> 0);
	};
	File.prototype.Fd = function() { return this.go$val.Fd(); };
	NewFile = go$pkg.NewFile = function(fd, name) {
		var fdi, f;
		fdi = (fd >> 0);
		if (fdi < 0) {
			return (go$ptrType(File)).nil;
		}
		f = new File.Ptr(new file.Ptr(fdi, name, (go$ptrType(dirInfo)).nil, 0));
		runtime.SetFinalizer(f.file, new (go$funcType([(go$ptrType(file))], [go$error], false))((function(recv) { return recv.close(); })));
		return f;
	};
	epipecheck = function(file$1, e) {
		var v, v$1;
		if (go$interfaceIsEqual(e, new syscall.Errno(32))) {
			if (atomic.AddInt32(new (go$ptrType(Go$Int32))(function() { return file$1.file.nepipe; }, function(v) { file$1.file.nepipe = v;; }), 1) >= 10) {
				sigpipe();
			}
		} else {
			atomic.StoreInt32(new (go$ptrType(Go$Int32))(function() { return file$1.file.nepipe; }, function(v$1) { file$1.file.nepipe = v$1;; }), 0);
		}
	};
	File.Ptr.prototype.Close = function() {
		var f;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			return go$pkg.ErrInvalid;
		}
		return f.file.close();
	};
	File.prototype.Close = function() { return this.go$val.Close(); };
	file.Ptr.prototype.close = function() {
		var file$1, err, e;
		file$1 = this;
		if (file$1 === (go$ptrType(file)).nil || file$1.fd < 0) {
			return new syscall.Errno(22);
		}
		err = null;
		e = syscall.Close(file$1.fd);
		if (!(go$interfaceIsEqual(e, null))) {
			err = new PathError.Ptr("close", file$1.name, e);
		}
		file$1.fd = -1;
		runtime.SetFinalizer(file$1, null);
		return err;
	};
	file.prototype.close = function() { return this.go$val.close(); };
	File.Ptr.prototype.Stat = function() {
		var fi, err, f, _tmp, _tmp$1, stat, _tmp$2, _tmp$3, _tmp$4, _tmp$5;
		fi = null;
		err = null;
		f = this;
		if (f === (go$ptrType(File)).nil) {
			_tmp = null; _tmp$1 = go$pkg.ErrInvalid; fi = _tmp; err = _tmp$1;
			return [fi, err];
		}
		stat = new syscall.Stat_t.Ptr();
		err = syscall.Fstat(f.file.fd, stat);
		if (!(go$interfaceIsEqual(err, null))) {
			_tmp$2 = null; _tmp$3 = new PathError.Ptr("stat", f.file.name, err); fi = _tmp$2; err = _tmp$3;
			return [fi, err];
		}
		_tmp$4 = fileInfoFromStat(stat, f.file.name); _tmp$5 = null; fi = _tmp$4; err = _tmp$5;
		return [fi, err];
	};
	File.prototype.Stat = function() { return this.go$val.Stat(); };
	Lstat = go$pkg.Lstat = function(name) {
		var fi, err, stat, _tmp, _tmp$1, _tmp$2, _tmp$3;
		fi = null;
		err = null;
		stat = new syscall.Stat_t.Ptr();
		err = syscall.Lstat(name, stat);
		if (!(go$interfaceIsEqual(err, null))) {
			_tmp = null; _tmp$1 = new PathError.Ptr("lstat", name, err); fi = _tmp; err = _tmp$1;
			return [fi, err];
		}
		_tmp$2 = fileInfoFromStat(stat, name); _tmp$3 = null; fi = _tmp$2; err = _tmp$3;
		return [fi, err];
	};
	File.Ptr.prototype.readdir = function(n) {
		var fi, err, f, dirname, _tuple, names, _ref, _i, filename, i, _tuple$1, fip, lerr, _tmp, _tmp$1;
		fi = (go$sliceType(FileInfo)).nil;
		err = null;
		f = this;
		dirname = f.file.name;
		if (dirname === "") {
			dirname = ".";
		}
		dirname = dirname + "/";
		_tuple = f.Readdirnames(n); names = _tuple[0]; err = _tuple[1];
		fi = (go$sliceType(FileInfo)).make(names.length, 0, function() { return null; });
		_ref = names;
		_i = 0;
		while (_i < _ref.length) {
			filename = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			_tuple$1 = lstat(dirname + filename); fip = _tuple$1[0]; lerr = _tuple$1[1];
			if (!(go$interfaceIsEqual(lerr, null))) {
				(i < 0 || i >= fi.length) ? go$throwRuntimeError("index out of range") : fi.array[fi.offset + i] = new fileStat.Ptr(filename, new Go$Int64(0, 0), 0, new time.Time.Ptr(), null);
				_i++;
				continue;
			}
			(i < 0 || i >= fi.length) ? go$throwRuntimeError("index out of range") : fi.array[fi.offset + i] = fip;
			_i++;
		}
		_tmp = fi; _tmp$1 = err; fi = _tmp; err = _tmp$1;
		return [fi, err];
	};
	File.prototype.readdir = function(n) { return this.go$val.readdir(n); };
	File.Ptr.prototype.read = function(b) {
		var n, err, f, _tuple;
		n = 0;
		err = null;
		f = this;
		_tuple = syscall.Read(f.file.fd, b); n = _tuple[0]; err = _tuple[1];
		return [n, err];
	};
	File.prototype.read = function(b) { return this.go$val.read(b); };
	File.Ptr.prototype.pread = function(b, off) {
		var n, err, f, _tuple;
		n = 0;
		err = null;
		f = this;
		_tuple = syscall.Pread(f.file.fd, b, off); n = _tuple[0]; err = _tuple[1];
		return [n, err];
	};
	File.prototype.pread = function(b, off) { return this.go$val.pread(b, off); };
	File.Ptr.prototype.write = function(b) {
		var n, err, f, _tuple, m, err$1, _tmp, _tmp$1;
		n = 0;
		err = null;
		f = this;
		while (true) {
			_tuple = syscall.Write(f.file.fd, b); m = _tuple[0]; err$1 = _tuple[1];
			n = n + (m) >> 0;
			if (0 < m && m < b.length || go$interfaceIsEqual(err$1, new syscall.Errno(4))) {
				b = go$subslice(b, m);
				continue;
			}
			_tmp = n; _tmp$1 = err$1; n = _tmp; err = _tmp$1;
			return [n, err];
		}
	};
	File.prototype.write = function(b) { return this.go$val.write(b); };
	File.Ptr.prototype.pwrite = function(b, off) {
		var n, err, f, _tuple;
		n = 0;
		err = null;
		f = this;
		_tuple = syscall.Pwrite(f.file.fd, b, off); n = _tuple[0]; err = _tuple[1];
		return [n, err];
	};
	File.prototype.pwrite = function(b, off) { return this.go$val.pwrite(b, off); };
	File.Ptr.prototype.seek = function(offset, whence) {
		var ret, err, f, _tuple;
		ret = new Go$Int64(0, 0);
		err = null;
		f = this;
		_tuple = syscall.Seek(f.file.fd, offset, whence); ret = _tuple[0]; err = _tuple[1];
		return [ret, err];
	};
	File.prototype.seek = function(offset, whence) { return this.go$val.seek(offset, whence); };
	basename = function(name) {
		var i;
		i = name.length - 1 >> 0;
		while (i > 0 && (name.charCodeAt(i) === 47)) {
			name = name.substring(0, i);
			i = i - 1 >> 0;
		}
		i = i - 1 >> 0;
		while (i >= 0) {
			if (name.charCodeAt(i) === 47) {
				name = name.substring((i + 1 >> 0));
				break;
			}
			i = i - 1 >> 0;
		}
		return name;
	};
	useSyscallwdDarwin = function(err) {
		return !(go$interfaceIsEqual(err, new syscall.Errno(45)));
	};
	fileInfoFromStat = function(st, name) {
		var _struct, _struct$1, fs, _ref;
		fs = new fileStat.Ptr(basename(name), st.Size, 0, (_struct$1 = timespecToTime((_struct = st.Mtimespec, new syscall.Timespec.Ptr(_struct.Sec, _struct.Nsec))), new time.Time.Ptr(_struct$1.sec, _struct$1.nsec, _struct$1.loc)), st);
		fs.mode = (((st.Mode & 511) >>> 0) >>> 0);
		_ref = (st.Mode & 61440) >>> 0;
		if (_ref === 24576 || _ref === 57344) {
			fs.mode = (fs.mode | 67108864) >>> 0;
		} else if (_ref === 8192) {
			fs.mode = (fs.mode | 69206016) >>> 0;
		} else if (_ref === 16384) {
			fs.mode = (fs.mode | 2147483648) >>> 0;
		} else if (_ref === 4096) {
			fs.mode = (fs.mode | 33554432) >>> 0;
		} else if (_ref === 40960) {
			fs.mode = (fs.mode | 134217728) >>> 0;
		} else if (_ref === 32768) {
		} else if (_ref === 49152) {
			fs.mode = (fs.mode | 16777216) >>> 0;
		}
		if (!((((st.Mode & 1024) >>> 0) === 0))) {
			fs.mode = (fs.mode | 4194304) >>> 0;
		}
		if (!((((st.Mode & 2048) >>> 0) === 0))) {
			fs.mode = (fs.mode | 8388608) >>> 0;
		}
		if (!((((st.Mode & 512) >>> 0) === 0))) {
			fs.mode = (fs.mode | 1048576) >>> 0;
		}
		return fs;
	};
	timespecToTime = function(ts) {
		var _struct;
		return (_struct = time.Unix(ts.Sec, ts.Nsec), new time.Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
	};
	FileMode.prototype.String = function() {
		var m, buf, w, _ref, _i, _rune, c, i, y, _ref$1, _i$1, _rune$1, c$1, i$1, y$1;
		m = this.go$val;
		buf = go$makeNativeArray("Uint8", 32, function() { return 0; });
		w = 0;
		_ref = "dalTLDpSugct";
		_i = 0;
		while (_i < _ref.length) {
			_rune = go$decodeRune(_ref, _i);
			c = _rune[0];
			i = _i;
			if (!((((m & (((y = ((31 - i >> 0) >>> 0), y < 32 ? (1 << y) : 0) >>> 0))) >>> 0) === 0))) {
				buf[w] = (c << 24 >>> 24);
				w = w + 1 >> 0;
			}
			_i += _rune[1];
		}
		if (w === 0) {
			buf[w] = 45;
			w = w + 1 >> 0;
		}
		_ref$1 = "rwxrwxrwx";
		_i$1 = 0;
		while (_i$1 < _ref$1.length) {
			_rune$1 = go$decodeRune(_ref$1, _i$1);
			c$1 = _rune$1[0];
			i$1 = _i$1;
			if (!((((m & (((y$1 = ((8 - i$1 >> 0) >>> 0), y$1 < 32 ? (1 << y$1) : 0) >>> 0))) >>> 0) === 0))) {
				buf[w] = (c$1 << 24 >>> 24);
			} else {
				buf[w] = 45;
			}
			w = w + 1 >> 0;
			_i$1 += _rune$1[1];
		}
		return go$bytesToString(go$subslice(new (go$sliceType(Go$Uint8))(buf), 0, w));
	};
	go$ptrType(FileMode).prototype.String = function() { return new FileMode(this.go$get()).String(); };
	FileMode.prototype.IsDir = function() {
		var m;
		m = this.go$val;
		return !((((m & 2147483648) >>> 0) === 0));
	};
	go$ptrType(FileMode).prototype.IsDir = function() { return new FileMode(this.go$get()).IsDir(); };
	FileMode.prototype.IsRegular = function() {
		var m;
		m = this.go$val;
		return ((m & 2399141888) >>> 0) === 0;
	};
	go$ptrType(FileMode).prototype.IsRegular = function() { return new FileMode(this.go$get()).IsRegular(); };
	FileMode.prototype.Perm = function() {
		var m;
		m = this.go$val;
		return (m & 511) >>> 0;
	};
	go$ptrType(FileMode).prototype.Perm = function() { return new FileMode(this.go$get()).Perm(); };
	fileStat.Ptr.prototype.Name = function() {
		var fs;
		fs = this;
		return fs.name;
	};
	fileStat.prototype.Name = function() { return this.go$val.Name(); };
	fileStat.Ptr.prototype.IsDir = function() {
		var fs;
		fs = this;
		return (new FileMode(fs.Mode())).IsDir();
	};
	fileStat.prototype.IsDir = function() { return this.go$val.IsDir(); };
	fileStat.Ptr.prototype.Size = function() {
		var fs;
		fs = this;
		return fs.size;
	};
	fileStat.prototype.Size = function() { return this.go$val.Size(); };
	fileStat.Ptr.prototype.Mode = function() {
		var fs;
		fs = this;
		return fs.mode;
	};
	fileStat.prototype.Mode = function() { return this.go$val.Mode(); };
	fileStat.Ptr.prototype.ModTime = function() {
		var fs, _struct;
		fs = this;
		return (_struct = fs.modTime, new time.Time.Ptr(_struct.sec, _struct.nsec, _struct.loc));
	};
	fileStat.prototype.ModTime = function() { return this.go$val.ModTime(); };
	fileStat.Ptr.prototype.Sys = function() {
		var fs;
		fs = this;
		return fs.sys;
	};
	fileStat.prototype.Sys = function() { return this.go$val.Sys(); };
	go$pkg.init = function() {
		(go$ptrType(PathError)).methods = [["Error", "", [], [Go$String], false, -1]];
		PathError.init([["Op", "Op", "", Go$String, ""], ["Path", "Path", "", Go$String, ""], ["Err", "Err", "", go$error, ""]]);
		(go$ptrType(SyscallError)).methods = [["Error", "", [], [Go$String], false, -1]];
		SyscallError.init([["Syscall", "Syscall", "", Go$String, ""], ["Err", "Err", "", go$error, ""]]);
		File.methods = [["close", "os", [], [go$error], false, 0]];
		(go$ptrType(File)).methods = [["Chdir", "", [], [go$error], false, -1], ["Chmod", "", [FileMode], [go$error], false, -1], ["Chown", "", [Go$Int, Go$Int], [go$error], false, -1], ["Close", "", [], [go$error], false, -1], ["Fd", "", [], [Go$Uintptr], false, -1], ["Name", "", [], [Go$String], false, -1], ["Read", "", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1], ["ReadAt", "", [(go$sliceType(Go$Uint8)), Go$Int64], [Go$Int, go$error], false, -1], ["Readdir", "", [Go$Int], [(go$sliceType(FileInfo)), go$error], false, -1], ["Readdirnames", "", [Go$Int], [(go$sliceType(Go$String)), go$error], false, -1], ["Seek", "", [Go$Int64, Go$Int], [Go$Int64, go$error], false, -1], ["Stat", "", [], [FileInfo, go$error], false, -1], ["Sync", "", [], [go$error], false, -1], ["Truncate", "", [Go$Int64], [go$error], false, -1], ["Write", "", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1], ["WriteAt", "", [(go$sliceType(Go$Uint8)), Go$Int64], [Go$Int, go$error], false, -1], ["WriteString", "", [Go$String], [Go$Int, go$error], false, -1], ["close", "os", [], [go$error], false, 0], ["pread", "os", [(go$sliceType(Go$Uint8)), Go$Int64], [Go$Int, go$error], false, -1], ["pwrite", "os", [(go$sliceType(Go$Uint8)), Go$Int64], [Go$Int, go$error], false, -1], ["read", "os", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1], ["readdir", "os", [Go$Int], [(go$sliceType(FileInfo)), go$error], false, -1], ["readdirnames", "os", [Go$Int], [(go$sliceType(Go$String)), go$error], false, -1], ["seek", "os", [Go$Int64, Go$Int], [Go$Int64, go$error], false, -1], ["write", "os", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1]];
		File.init([["file", "", "os", (go$ptrType(file)), ""]]);
		(go$ptrType(file)).methods = [["close", "os", [], [go$error], false, -1]];
		file.init([["fd", "fd", "os", Go$Int, ""], ["name", "name", "os", Go$String, ""], ["dirinfo", "dirinfo", "os", (go$ptrType(dirInfo)), ""], ["nepipe", "nepipe", "os", Go$Int32, ""]]);
		dirInfo.init([["buf", "buf", "os", (go$sliceType(Go$Uint8)), ""], ["nbuf", "nbuf", "os", Go$Int, ""], ["bufp", "bufp", "os", Go$Int, ""]]);
		FileInfo.init([["IsDir", "", (go$funcType([], [Go$Bool], false))], ["ModTime", "", (go$funcType([], [time.Time], false))], ["Mode", "", (go$funcType([], [FileMode], false))], ["Name", "", (go$funcType([], [Go$String], false))], ["Size", "", (go$funcType([], [Go$Int64], false))], ["Sys", "", (go$funcType([], [go$emptyInterface], false))]]);
		FileMode.methods = [["IsDir", "", [], [Go$Bool], false, -1], ["IsRegular", "", [], [Go$Bool], false, -1], ["Perm", "", [], [FileMode], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(FileMode)).methods = [["IsDir", "", [], [Go$Bool], false, -1], ["IsRegular", "", [], [Go$Bool], false, -1], ["Perm", "", [], [FileMode], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(fileStat)).methods = [["IsDir", "", [], [Go$Bool], false, -1], ["ModTime", "", [], [time.Time], false, -1], ["Mode", "", [], [FileMode], false, -1], ["Name", "", [], [Go$String], false, -1], ["Size", "", [], [Go$Int64], false, -1], ["Sys", "", [], [go$emptyInterface], false, -1]];
		fileStat.init([["name", "name", "os", Go$String, ""], ["size", "size", "os", Go$Int64, ""], ["mode", "mode", "os", FileMode, ""], ["modTime", "modTime", "os", time.Time, ""], ["sys", "sys", "os", go$emptyInterface, ""]]);
		go$pkg.Args = (go$sliceType(Go$String)).nil;
		go$pkg.ErrInvalid = errors.New("invalid argument");
		go$pkg.ErrPermission = errors.New("permission denied");
		go$pkg.ErrExist = errors.New("file already exists");
		go$pkg.ErrNotExist = errors.New("file does not exist");
		go$pkg.Stdin = NewFile((syscall.Stdin >>> 0), "/dev/stdin");
		go$pkg.Stdout = NewFile((syscall.Stdout >>> 0), "/dev/stdout");
		go$pkg.Stderr = NewFile((syscall.Stderr >>> 0), "/dev/stderr");
		lstat = Lstat;
		useSyscallwd = (function() {
			return true;
		});
		var process, args, i;
		process = go$global.process;
		if (!(process === undefined)) {
			args = process.argv;
			go$pkg.Args = (go$sliceType(Go$String)).make((go$parseInt(args.length) - 1 >> 0), 0, function() { return ""; });
			i = 0;
			while (i < (go$parseInt(args.length) - 1 >> 0)) {
				(i < 0 || i >= go$pkg.Args.length) ? go$throwRuntimeError("index out of range") : go$pkg.Args.array[go$pkg.Args.offset + i] = go$internalize(args[(i + 1 >> 0)], Go$String);
				i = i + 1 >> 0;
			}
		}
		useSyscallwd = useSyscallwdDarwin;
	}
	return go$pkg;
})();
go$packages["strconv"] = (function() {
	var go$pkg = {}, math = go$packages["math"], errors = go$packages["errors"], utf8 = go$packages["unicode/utf8"], decimal, leftCheat, extFloat, floatInfo, decimalSlice, digitZero, trim, rightShift, prefixIsLessThan, leftShift, shouldRoundUp, frexp10Many, adjustLastDigitFixed, adjustLastDigit, AppendFloat, genericFtoa, bigFtoa, formatDigits, roundShortest, fmtE, fmtF, fmtB, max, FormatInt, Itoa, formatBits, quoteWith, Quote, QuoteToASCII, QuoteRune, AppendQuoteRune, QuoteRuneToASCII, AppendQuoteRuneToASCII, CanBackquote, unhex, UnquoteChar, Unquote, contains, bsearch16, bsearch32, IsPrint, optimize, leftcheats, smallPowersOfTen, powersOfTen, uint64pow10, float32info, float64info, isPrint16, isNotPrint16, isPrint32, isNotPrint32, shifts;
	decimal = go$pkg.decimal = go$newType(0, "Struct", "strconv.decimal", "decimal", "strconv", function(d_, nd_, dp_, neg_, trunc_) {
		this.go$val = this;
		this.d = d_ !== undefined ? d_ : go$makeNativeArray("Uint8", 800, function() { return 0; });
		this.nd = nd_ !== undefined ? nd_ : 0;
		this.dp = dp_ !== undefined ? dp_ : 0;
		this.neg = neg_ !== undefined ? neg_ : false;
		this.trunc = trunc_ !== undefined ? trunc_ : false;
	});
	leftCheat = go$pkg.leftCheat = go$newType(0, "Struct", "strconv.leftCheat", "leftCheat", "strconv", function(delta_, cutoff_) {
		this.go$val = this;
		this.delta = delta_ !== undefined ? delta_ : 0;
		this.cutoff = cutoff_ !== undefined ? cutoff_ : "";
	});
	extFloat = go$pkg.extFloat = go$newType(0, "Struct", "strconv.extFloat", "extFloat", "strconv", function(mant_, exp_, neg_) {
		this.go$val = this;
		this.mant = mant_ !== undefined ? mant_ : new Go$Uint64(0, 0);
		this.exp = exp_ !== undefined ? exp_ : 0;
		this.neg = neg_ !== undefined ? neg_ : false;
	});
	floatInfo = go$pkg.floatInfo = go$newType(0, "Struct", "strconv.floatInfo", "floatInfo", "strconv", function(mantbits_, expbits_, bias_) {
		this.go$val = this;
		this.mantbits = mantbits_ !== undefined ? mantbits_ : 0;
		this.expbits = expbits_ !== undefined ? expbits_ : 0;
		this.bias = bias_ !== undefined ? bias_ : 0;
	});
	decimalSlice = go$pkg.decimalSlice = go$newType(0, "Struct", "strconv.decimalSlice", "decimalSlice", "strconv", function(d_, nd_, dp_, neg_) {
		this.go$val = this;
		this.d = d_ !== undefined ? d_ : (go$sliceType(Go$Uint8)).nil;
		this.nd = nd_ !== undefined ? nd_ : 0;
		this.dp = dp_ !== undefined ? dp_ : 0;
		this.neg = neg_ !== undefined ? neg_ : false;
	});
	decimal.Ptr.prototype.String = function() {
		var a, n, buf, w;
		a = this;
		n = 10 + a.nd >> 0;
		if (a.dp > 0) {
			n = n + (a.dp) >> 0;
		}
		if (a.dp < 0) {
			n = n + (-a.dp) >> 0;
		}
		buf = (go$sliceType(Go$Uint8)).make(n, 0, function() { return 0; });
		w = 0;
		if (a.nd === 0) {
			return "0";
		} else if (a.dp <= 0) {
			(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = 48;
			w = w + 1 >> 0;
			(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = 46;
			w = w + 1 >> 0;
			w = w + (digitZero(go$subslice(buf, w, (w + -a.dp >> 0)))) >> 0;
			w = w + (go$copySlice(go$subslice(buf, w), go$subslice(new (go$sliceType(Go$Uint8))(a.d), 0, a.nd))) >> 0;
		} else if (a.dp < a.nd) {
			w = w + (go$copySlice(go$subslice(buf, w), go$subslice(new (go$sliceType(Go$Uint8))(a.d), 0, a.dp))) >> 0;
			(w < 0 || w >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + w] = 46;
			w = w + 1 >> 0;
			w = w + (go$copySlice(go$subslice(buf, w), go$subslice(new (go$sliceType(Go$Uint8))(a.d), a.dp, a.nd))) >> 0;
		} else {
			w = w + (go$copySlice(go$subslice(buf, w), go$subslice(new (go$sliceType(Go$Uint8))(a.d), 0, a.nd))) >> 0;
			w = w + (digitZero(go$subslice(buf, w, ((w + a.dp >> 0) - a.nd >> 0)))) >> 0;
		}
		return go$bytesToString(go$subslice(buf, 0, w));
	};
	decimal.prototype.String = function() { return this.go$val.String(); };
	digitZero = function(dst) {
		var _ref, _i, i;
		_ref = dst;
		_i = 0;
		while (_i < _ref.length) {
			i = _i;
			(i < 0 || i >= dst.length) ? go$throwRuntimeError("index out of range") : dst.array[dst.offset + i] = 48;
			_i++;
		}
		return dst.length;
	};
	trim = function(a) {
		while (a.nd > 0 && (a.d[(a.nd - 1 >> 0)] === 48)) {
			a.nd = a.nd - 1 >> 0;
		}
		if (a.nd === 0) {
			a.dp = 0;
		}
	};
	decimal.Ptr.prototype.Assign = function(v) {
		var a, buf, n, v1, x;
		a = this;
		buf = go$makeNativeArray("Uint8", 24, function() { return 0; });
		n = 0;
		while ((v.high > 0 || (v.high === 0 && v.low > 0))) {
			v1 = go$div64(v, new Go$Uint64(0, 10), false);
			v = (x = go$mul64(new Go$Uint64(0, 10), v1), new Go$Uint64(v.high - x.high, v.low - x.low));
			buf[n] = (new Go$Uint64(v.high + 0, v.low + 48).low << 24 >>> 24);
			n = n + 1 >> 0;
			v = v1;
		}
		a.nd = 0;
		n = n - 1 >> 0;
		while (n >= 0) {
			a.d[a.nd] = buf[n];
			a.nd = a.nd + 1 >> 0;
			n = n - 1 >> 0;
		}
		a.dp = a.nd;
		trim(a);
	};
	decimal.prototype.Assign = function(v) { return this.go$val.Assign(v); };
	rightShift = function(a, k) {
		var r, w, n, c, c$1, dig, y, dig$1, y$1;
		r = 0;
		w = 0;
		n = 0;
		while (((n >> go$min(k, 31)) >> 0) === 0) {
			if (r >= a.nd) {
				if (n === 0) {
					a.nd = 0;
					return;
				}
				while (((n >> go$min(k, 31)) >> 0) === 0) {
					n = (((n >>> 16 << 16) * 10 >> 0) + (n << 16 >>> 16) * 10) >> 0;
					r = r + 1 >> 0;
				}
				break;
			}
			c = (a.d[r] >> 0);
			n = (((((n >>> 16 << 16) * 10 >> 0) + (n << 16 >>> 16) * 10) >> 0) + c >> 0) - 48 >> 0;
			r = r + 1 >> 0;
		}
		a.dp = a.dp - ((r - 1 >> 0)) >> 0;
		while (r < a.nd) {
			c$1 = (a.d[r] >> 0);
			dig = (n >> go$min(k, 31)) >> 0;
			n = n - (((y = k, y < 32 ? (dig << y) : 0) >> 0)) >> 0;
			a.d[w] = ((dig + 48 >> 0) << 24 >>> 24);
			w = w + 1 >> 0;
			n = (((((n >>> 16 << 16) * 10 >> 0) + (n << 16 >>> 16) * 10) >> 0) + c$1 >> 0) - 48 >> 0;
			r = r + 1 >> 0;
		}
		while (n > 0) {
			dig$1 = (n >> go$min(k, 31)) >> 0;
			n = n - (((y$1 = k, y$1 < 32 ? (dig$1 << y$1) : 0) >> 0)) >> 0;
			if (w < 800) {
				a.d[w] = ((dig$1 + 48 >> 0) << 24 >>> 24);
				w = w + 1 >> 0;
			} else if (dig$1 > 0) {
				a.trunc = true;
			}
			n = (((n >>> 16 << 16) * 10 >> 0) + (n << 16 >>> 16) * 10) >> 0;
		}
		a.nd = w;
		trim(a);
	};
	prefixIsLessThan = function(b, s) {
		var i;
		i = 0;
		while (i < s.length) {
			if (i >= b.length) {
				return true;
			}
			if (!((((i < 0 || i >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + i]) === s.charCodeAt(i)))) {
				return ((i < 0 || i >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + i]) < s.charCodeAt(i);
			}
			i = i + 1 >> 0;
		}
		return false;
	};
	leftShift = function(a, k) {
		var delta, r, w, n, y, _q, quo, rem, _q$1, quo$1, rem$1;
		delta = ((k < 0 || k >= leftcheats.length) ? go$throwRuntimeError("index out of range") : leftcheats.array[leftcheats.offset + k]).delta;
		if (prefixIsLessThan(go$subslice(new (go$sliceType(Go$Uint8))(a.d), 0, a.nd), ((k < 0 || k >= leftcheats.length) ? go$throwRuntimeError("index out of range") : leftcheats.array[leftcheats.offset + k]).cutoff)) {
			delta = delta - 1 >> 0;
		}
		r = a.nd;
		w = a.nd + delta >> 0;
		n = 0;
		r = r - 1 >> 0;
		while (r >= 0) {
			n = n + (((y = k, y < 32 ? ((((a.d[r] >> 0) - 48 >> 0)) << y) : 0) >> 0)) >> 0;
			quo = (_q = n / 10, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
			rem = n - ((((10 >>> 16 << 16) * quo >> 0) + (10 << 16 >>> 16) * quo) >> 0) >> 0;
			w = w - 1 >> 0;
			if (w < 800) {
				a.d[w] = ((rem + 48 >> 0) << 24 >>> 24);
			} else if (!((rem === 0))) {
				a.trunc = true;
			}
			n = quo;
			r = r - 1 >> 0;
		}
		while (n > 0) {
			quo$1 = (_q$1 = n / 10, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >> 0 : go$throwRuntimeError("integer divide by zero"));
			rem$1 = n - ((((10 >>> 16 << 16) * quo$1 >> 0) + (10 << 16 >>> 16) * quo$1) >> 0) >> 0;
			w = w - 1 >> 0;
			if (w < 800) {
				a.d[w] = ((rem$1 + 48 >> 0) << 24 >>> 24);
			} else if (!((rem$1 === 0))) {
				a.trunc = true;
			}
			n = quo$1;
		}
		a.nd = a.nd + (delta) >> 0;
		if (a.nd >= 800) {
			a.nd = 800;
		}
		a.dp = a.dp + (delta) >> 0;
		trim(a);
	};
	decimal.Ptr.prototype.Shift = function(k) {
		var a;
		a = this;
		if (a.nd === 0) {
		} else if (k > 0) {
			while (k > 27) {
				leftShift(a, 27);
				k = k - 27 >> 0;
			}
			leftShift(a, (k >>> 0));
		} else if (k < 0) {
			while (k < -27) {
				rightShift(a, 27);
				k = k + 27 >> 0;
			}
			rightShift(a, (-k >>> 0));
		}
	};
	decimal.prototype.Shift = function(k) { return this.go$val.Shift(k); };
	shouldRoundUp = function(a, nd) {
		var _r;
		if (nd < 0 || nd >= a.nd) {
			return false;
		}
		if ((a.d[nd] === 53) && ((nd + 1 >> 0) === a.nd)) {
			if (a.trunc) {
				return true;
			}
			return nd > 0 && !(((_r = ((a.d[(nd - 1 >> 0)] - 48 << 24 >>> 24)) % 2, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) === 0));
		}
		return a.d[nd] >= 53;
	};
	decimal.Ptr.prototype.Round = function(nd) {
		var a;
		a = this;
		if (nd < 0 || nd >= a.nd) {
			return;
		}
		if (shouldRoundUp(a, nd)) {
			a.RoundUp(nd);
		} else {
			a.RoundDown(nd);
		}
	};
	decimal.prototype.Round = function(nd) { return this.go$val.Round(nd); };
	decimal.Ptr.prototype.RoundDown = function(nd) {
		var a;
		a = this;
		if (nd < 0 || nd >= a.nd) {
			return;
		}
		a.nd = nd;
		trim(a);
	};
	decimal.prototype.RoundDown = function(nd) { return this.go$val.RoundDown(nd); };
	decimal.Ptr.prototype.RoundUp = function(nd) {
		var a, i, c, _lhs, _index;
		a = this;
		if (nd < 0 || nd >= a.nd) {
			return;
		}
		i = nd - 1 >> 0;
		while (i >= 0) {
			c = a.d[i];
			if (c < 57) {
				_lhs = a.d; _index = i; _lhs[_index] = _lhs[_index] + 1 << 24 >>> 24;
				a.nd = i + 1 >> 0;
				return;
			}
			i = i - 1 >> 0;
		}
		a.d[0] = 49;
		a.nd = 1;
		a.dp = a.dp + 1 >> 0;
	};
	decimal.prototype.RoundUp = function(nd) { return this.go$val.RoundUp(nd); };
	decimal.Ptr.prototype.RoundedInteger = function() {
		var a, i, n, x, x$1;
		a = this;
		if (a.dp > 20) {
			return new Go$Uint64(4294967295, 4294967295);
		}
		i = 0;
		n = new Go$Uint64(0, 0);
		i = 0;
		while (i < a.dp && i < a.nd) {
			n = (x = go$mul64(n, new Go$Uint64(0, 10)), x$1 = new Go$Uint64(0, (a.d[i] - 48 << 24 >>> 24)), new Go$Uint64(x.high + x$1.high, x.low + x$1.low));
			i = i + 1 >> 0;
		}
		while (i < a.dp) {
			n = go$mul64(n, new Go$Uint64(0, 10));
			i = i + 1 >> 0;
		}
		if (shouldRoundUp(a, a.dp)) {
			n = new Go$Uint64(n.high + 0, n.low + 1);
		}
		return n;
	};
	decimal.prototype.RoundedInteger = function() { return this.go$val.RoundedInteger(); };
	extFloat.Ptr.prototype.AssignComputeBounds = function(mant, exp, neg, flt) {
		var lower, upper, f, x, _tmp, _struct, _tmp$1, _struct$1, _struct$2, _struct$3, expBiased, x$1, x$2, x$3, x$4, _struct$4, _struct$5;
		lower = new extFloat.Ptr();
		upper = new extFloat.Ptr();
		f = this;
		f.mant = mant;
		f.exp = exp - (flt.mantbits >> 0) >> 0;
		f.neg = neg;
		if (f.exp <= 0 && (x = go$shiftLeft64((go$shiftRightUint64(mant, (-f.exp >>> 0))), (-f.exp >>> 0)), (mant.high === x.high && mant.low === x.low))) {
			f.mant = go$shiftRightUint64(f.mant, ((-f.exp >>> 0)));
			f.exp = 0;
			_tmp = (_struct = f, new extFloat.Ptr(_struct.mant, _struct.exp, _struct.neg)); _tmp$1 = (_struct$1 = f, new extFloat.Ptr(_struct$1.mant, _struct$1.exp, _struct$1.neg)); lower = _tmp; upper = _tmp$1;
			return [(_struct$2 = lower, new extFloat.Ptr(_struct$2.mant, _struct$2.exp, _struct$2.neg)), (_struct$3 = upper, new extFloat.Ptr(_struct$3.mant, _struct$3.exp, _struct$3.neg))];
		}
		expBiased = exp - flt.bias >> 0;
		upper = new extFloat.Ptr((x$1 = go$mul64(new Go$Uint64(0, 2), f.mant), new Go$Uint64(x$1.high + 0, x$1.low + 1)), f.exp - 1 >> 0, f.neg);
		if (!((x$2 = go$shiftLeft64(new Go$Uint64(0, 1), flt.mantbits), (mant.high === x$2.high && mant.low === x$2.low))) || (expBiased === 1)) {
			lower = new extFloat.Ptr((x$3 = go$mul64(new Go$Uint64(0, 2), f.mant), new Go$Uint64(x$3.high - 0, x$3.low - 1)), f.exp - 1 >> 0, f.neg);
		} else {
			lower = new extFloat.Ptr((x$4 = go$mul64(new Go$Uint64(0, 4), f.mant), new Go$Uint64(x$4.high - 0, x$4.low - 1)), f.exp - 2 >> 0, f.neg);
		}
		return [(_struct$4 = lower, new extFloat.Ptr(_struct$4.mant, _struct$4.exp, _struct$4.neg)), (_struct$5 = upper, new extFloat.Ptr(_struct$5.mant, _struct$5.exp, _struct$5.neg))];
	};
	extFloat.prototype.AssignComputeBounds = function(mant, exp, neg, flt) { return this.go$val.AssignComputeBounds(mant, exp, neg, flt); };
	extFloat.Ptr.prototype.Normalize = function() {
		var shift, f, _tmp, _tmp$1, mant, exp, x, x$1, x$2, x$3, x$4, x$5, _tmp$2, _tmp$3;
		shift = 0;
		f = this;
		_tmp = f.mant; _tmp$1 = f.exp; mant = _tmp; exp = _tmp$1;
		if ((mant.high === 0 && mant.low === 0)) {
			shift = 0;
			return shift;
		}
		if ((x = go$shiftRightUint64(mant, 32), (x.high === 0 && x.low === 0))) {
			mant = go$shiftLeft64(mant, 32);
			exp = exp - 32 >> 0;
		}
		if ((x$1 = go$shiftRightUint64(mant, 48), (x$1.high === 0 && x$1.low === 0))) {
			mant = go$shiftLeft64(mant, 16);
			exp = exp - 16 >> 0;
		}
		if ((x$2 = go$shiftRightUint64(mant, 56), (x$2.high === 0 && x$2.low === 0))) {
			mant = go$shiftLeft64(mant, 8);
			exp = exp - 8 >> 0;
		}
		if ((x$3 = go$shiftRightUint64(mant, 60), (x$3.high === 0 && x$3.low === 0))) {
			mant = go$shiftLeft64(mant, 4);
			exp = exp - 4 >> 0;
		}
		if ((x$4 = go$shiftRightUint64(mant, 62), (x$4.high === 0 && x$4.low === 0))) {
			mant = go$shiftLeft64(mant, 2);
			exp = exp - 2 >> 0;
		}
		if ((x$5 = go$shiftRightUint64(mant, 63), (x$5.high === 0 && x$5.low === 0))) {
			mant = go$shiftLeft64(mant, 1);
			exp = exp - 1 >> 0;
		}
		shift = ((f.exp - exp >> 0) >>> 0);
		_tmp$2 = mant; _tmp$3 = exp; f.mant = _tmp$2; f.exp = _tmp$3;
		return shift;
	};
	extFloat.prototype.Normalize = function() { return this.go$val.Normalize(); };
	extFloat.Ptr.prototype.Multiply = function(g) {
		var f, _tmp, _tmp$1, fhi, flo, _tmp$2, _tmp$3, ghi, glo, cross1, cross2, x, x$1, x$2, x$3, x$4, x$5, x$6, x$7, rem, x$8, x$9;
		f = this;
		_tmp = go$shiftRightUint64(f.mant, 32); _tmp$1 = new Go$Uint64(0, (f.mant.low >>> 0)); fhi = _tmp; flo = _tmp$1;
		_tmp$2 = go$shiftRightUint64(g.mant, 32); _tmp$3 = new Go$Uint64(0, (g.mant.low >>> 0)); ghi = _tmp$2; glo = _tmp$3;
		cross1 = go$mul64(fhi, glo);
		cross2 = go$mul64(flo, ghi);
		f.mant = (x = (x$1 = go$mul64(fhi, ghi), x$2 = go$shiftRightUint64(cross1, 32), new Go$Uint64(x$1.high + x$2.high, x$1.low + x$2.low)), x$3 = go$shiftRightUint64(cross2, 32), new Go$Uint64(x.high + x$3.high, x.low + x$3.low));
		rem = (x$4 = (x$5 = new Go$Uint64(0, (cross1.low >>> 0)), x$6 = new Go$Uint64(0, (cross2.low >>> 0)), new Go$Uint64(x$5.high + x$6.high, x$5.low + x$6.low)), x$7 = go$shiftRightUint64((go$mul64(flo, glo)), 32), new Go$Uint64(x$4.high + x$7.high, x$4.low + x$7.low));
		rem = new Go$Uint64(rem.high + 0, rem.low + 2147483648);
		f.mant = (x$8 = f.mant, x$9 = (go$shiftRightUint64(rem, 32)), new Go$Uint64(x$8.high + x$9.high, x$8.low + x$9.low));
		f.exp = (f.exp + g.exp >> 0) + 64 >> 0;
	};
	extFloat.prototype.Multiply = function(g) { return this.go$val.Multiply(g); };
	extFloat.Ptr.prototype.AssignDecimal = function(mantissa, exp10, neg, trunc, flt) {
		var ok, f, errors$1, _q, i, _r, adjExp, x, _struct, _struct$1, shift, y, denormalExp, extrabits, halfway, x$1, x$2, x$3, mant_extra, x$4, x$5, x$6, x$7, x$8, x$9, x$10, x$11;
		ok = false;
		f = this;
		errors$1 = 0;
		if (trunc) {
			errors$1 = errors$1 + 4 >> 0;
		}
		f.mant = mantissa;
		f.exp = 0;
		f.neg = neg;
		i = (_q = ((exp10 - -348 >> 0)) / 8, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		if (exp10 < -348 || i >= 87) {
			ok = false;
			return ok;
		}
		adjExp = (_r = ((exp10 - -348 >> 0)) % 8, _r === _r ? _r : go$throwRuntimeError("integer divide by zero"));
		if (adjExp < 19 && (x = uint64pow10[(19 - adjExp >> 0)], (mantissa.high < x.high || (mantissa.high === x.high && mantissa.low < x.low)))) {
			f.mant = go$mul64(f.mant, (uint64pow10[adjExp]));
			f.Normalize();
		} else {
			f.Normalize();
			f.Multiply((_struct = smallPowersOfTen[adjExp], new extFloat.Ptr(_struct.mant, _struct.exp, _struct.neg)));
			errors$1 = errors$1 + 4 >> 0;
		}
		f.Multiply((_struct$1 = powersOfTen[i], new extFloat.Ptr(_struct$1.mant, _struct$1.exp, _struct$1.neg)));
		if (errors$1 > 0) {
			errors$1 = errors$1 + 1 >> 0;
		}
		errors$1 = errors$1 + 4 >> 0;
		shift = f.Normalize();
		errors$1 = (y = (shift), y < 32 ? (errors$1 << y) : 0) >> 0;
		denormalExp = flt.bias - 63 >> 0;
		extrabits = 0;
		if (f.exp <= denormalExp) {
			extrabits = (((63 - flt.mantbits >>> 0) + 1 >>> 0) + ((denormalExp - f.exp >> 0) >>> 0) >>> 0);
		} else {
			extrabits = (63 - flt.mantbits >>> 0);
		}
		halfway = go$shiftLeft64(new Go$Uint64(0, 1), ((extrabits - 1 >>> 0)));
		mant_extra = (x$1 = f.mant, x$2 = (x$3 = go$shiftLeft64(new Go$Uint64(0, 1), extrabits), new Go$Uint64(x$3.high - 0, x$3.low - 1)), new Go$Uint64(x$1.high & x$2.high, (x$1.low & x$2.low) >>> 0));
		if ((x$4 = (x$5 = new Go$Int64(halfway.high, halfway.low), x$6 = new Go$Int64(0, errors$1), new Go$Int64(x$5.high - x$6.high, x$5.low - x$6.low)), x$7 = new Go$Int64(mant_extra.high, mant_extra.low), (x$4.high < x$7.high || (x$4.high === x$7.high && x$4.low < x$7.low))) && (x$8 = new Go$Int64(mant_extra.high, mant_extra.low), x$9 = (x$10 = new Go$Int64(halfway.high, halfway.low), x$11 = new Go$Int64(0, errors$1), new Go$Int64(x$10.high + x$11.high, x$10.low + x$11.low)), (x$8.high < x$9.high || (x$8.high === x$9.high && x$8.low < x$9.low)))) {
			ok = false;
			return ok;
		}
		ok = true;
		return ok;
	};
	extFloat.prototype.AssignDecimal = function(mantissa, exp10, neg, trunc, flt) { return this.go$val.AssignDecimal(mantissa, exp10, neg, trunc, flt); };
	extFloat.Ptr.prototype.frexp10 = function() {
		var exp10, index, f, _q, x, approxExp10, _q$1, i, exp, _struct, _tmp, _tmp$1;
		exp10 = 0;
		index = 0;
		f = this;
		approxExp10 = (_q = (x = (-46 - f.exp >> 0), (((x >>> 16 << 16) * 28 >> 0) + (x << 16 >>> 16) * 28) >> 0) / 93, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		i = (_q$1 = ((approxExp10 - -348 >> 0)) / 8, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >> 0 : go$throwRuntimeError("integer divide by zero"));
		Loop:
		while (true) {
			exp = (f.exp + powersOfTen[i].exp >> 0) + 64 >> 0;
			if (exp < -60) {
				i = i + 1 >> 0;
			} else if (exp > -32) {
				i = i - 1 >> 0;
			} else {
				break Loop;
			}
		}
		f.Multiply((_struct = powersOfTen[i], new extFloat.Ptr(_struct.mant, _struct.exp, _struct.neg)));
		_tmp = -((-348 + ((((i >>> 16 << 16) * 8 >> 0) + (i << 16 >>> 16) * 8) >> 0) >> 0)); _tmp$1 = i; exp10 = _tmp; index = _tmp$1;
		return [exp10, index];
	};
	extFloat.prototype.frexp10 = function() { return this.go$val.frexp10(); };
	frexp10Many = function(a, b, c) {
		var exp10, _tuple, i, _struct, _struct$1;
		exp10 = 0;
		_tuple = c.frexp10(); exp10 = _tuple[0]; i = _tuple[1];
		a.Multiply((_struct = powersOfTen[i], new extFloat.Ptr(_struct.mant, _struct.exp, _struct.neg)));
		b.Multiply((_struct$1 = powersOfTen[i], new extFloat.Ptr(_struct$1.mant, _struct$1.exp, _struct$1.neg)));
		return exp10;
	};
	extFloat.Ptr.prototype.FixedDecimal = function(d, n) {
		var f, x, _tuple, exp10, shift, integer, x$1, x$2, fraction, nonAsciiName, needed, integerDigits, pow10, _tmp, _tmp$1, i, pow, x$3, rest, _q, x$4, buf, pos, v, _q$1, v1, i$1, x$5, x$6, nd, x$7, x$8, digit, x$9, x$10, x$11, ok, i$2, x$12;
		f = this;
		if ((x = f.mant, (x.high === 0 && x.low === 0))) {
			d.nd = 0;
			d.dp = 0;
			d.neg = f.neg;
			return true;
		}
		if (n === 0) {
			throw go$panic(new Go$String("strconv: internal error: extFloat.FixedDecimal called with n == 0"));
		}
		f.Normalize();
		_tuple = f.frexp10(); exp10 = _tuple[0];
		shift = (-f.exp >>> 0);
		integer = (go$shiftRightUint64(f.mant, shift).low >>> 0);
		fraction = (x$1 = f.mant, x$2 = go$shiftLeft64(new Go$Uint64(0, integer), shift), new Go$Uint64(x$1.high - x$2.high, x$1.low - x$2.low));
		nonAsciiName = new Go$Uint64(0, 1);
		needed = n;
		integerDigits = 0;
		pow10 = new Go$Uint64(0, 1);
		_tmp = 0; _tmp$1 = new Go$Uint64(0, 1); i = _tmp; pow = _tmp$1;
		while (i < 20) {
			if ((x$3 = new Go$Uint64(0, integer), (pow.high > x$3.high || (pow.high === x$3.high && pow.low > x$3.low)))) {
				integerDigits = i;
				break;
			}
			pow = go$mul64(pow, new Go$Uint64(0, 10));
			i = i + 1 >> 0;
		}
		rest = integer;
		if (integerDigits > needed) {
			pow10 = uint64pow10[(integerDigits - needed >> 0)];
			integer = (_q = integer / ((pow10.low >>> 0)), (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >>> 0 : go$throwRuntimeError("integer divide by zero"));
			rest = rest - ((x$4 = (pow10.low >>> 0), (((integer >>> 16 << 16) * x$4 >>> 0) + (integer << 16 >>> 16) * x$4) >>> 0)) >>> 0;
		} else {
			rest = 0;
		}
		buf = go$makeNativeArray("Uint8", 32, function() { return 0; });
		pos = 32;
		v = integer;
		while (v > 0) {
			v1 = (_q$1 = v / 10, (_q$1 === _q$1 && _q$1 !== 1/0 && _q$1 !== -1/0) ? _q$1 >>> 0 : go$throwRuntimeError("integer divide by zero"));
			v = v - (((((10 >>> 16 << 16) * v1 >>> 0) + (10 << 16 >>> 16) * v1) >>> 0)) >>> 0;
			pos = pos - 1 >> 0;
			buf[pos] = ((v + 48 >>> 0) << 24 >>> 24);
			v = v1;
		}
		i$1 = pos;
		while (i$1 < 32) {
			(x$5 = d.d, x$6 = i$1 - pos >> 0, (x$6 < 0 || x$6 >= x$5.length) ? go$throwRuntimeError("index out of range") : x$5.array[x$5.offset + x$6] = buf[i$1]);
			i$1 = i$1 + 1 >> 0;
		}
		nd = 32 - pos >> 0;
		d.nd = nd;
		d.dp = integerDigits + exp10 >> 0;
		needed = needed - (nd) >> 0;
		if (needed > 0) {
			if (!((rest === 0)) || !((pow10.high === 0 && pow10.low === 1))) {
				throw go$panic(new Go$String("strconv: internal error, rest != 0 but needed > 0"));
			}
			while (needed > 0) {
				fraction = go$mul64(fraction, new Go$Uint64(0, 10));
				nonAsciiName = go$mul64(nonAsciiName, new Go$Uint64(0, 10));
				if ((x$7 = go$mul64(new Go$Uint64(0, 2), nonAsciiName), x$8 = go$shiftLeft64(new Go$Uint64(0, 1), shift), (x$7.high > x$8.high || (x$7.high === x$8.high && x$7.low > x$8.low)))) {
					return false;
				}
				digit = go$shiftRightUint64(fraction, shift);
				(x$9 = d.d, (nd < 0 || nd >= x$9.length) ? go$throwRuntimeError("index out of range") : x$9.array[x$9.offset + nd] = (new Go$Uint64(digit.high + 0, digit.low + 48).low << 24 >>> 24));
				fraction = (x$10 = go$shiftLeft64(digit, shift), new Go$Uint64(fraction.high - x$10.high, fraction.low - x$10.low));
				nd = nd + 1 >> 0;
				needed = needed - 1 >> 0;
			}
			d.nd = nd;
		}
		ok = adjustLastDigitFixed(d, (x$11 = go$shiftLeft64(new Go$Uint64(0, rest), shift), new Go$Uint64(x$11.high | fraction.high, (x$11.low | fraction.low) >>> 0)), pow10, shift, nonAsciiName);
		if (!ok) {
			return false;
		}
		i$2 = d.nd - 1 >> 0;
		while (i$2 >= 0) {
			if (!(((x$12 = d.d, ((i$2 < 0 || i$2 >= x$12.length) ? go$throwRuntimeError("index out of range") : x$12.array[x$12.offset + i$2])) === 48))) {
				d.nd = i$2 + 1 >> 0;
				break;
			}
			i$2 = i$2 - 1 >> 0;
		}
		return true;
	};
	extFloat.prototype.FixedDecimal = function(d, n) { return this.go$val.FixedDecimal(d, n); };
	adjustLastDigitFixed = function(d, num, den, shift, nonAsciiName) {
		var x, x$1, x$2, x$3, x$4, x$5, x$6, i, x$7, x$8, _lhs, _index;
		if ((x = go$shiftLeft64(den, shift), (num.high > x.high || (num.high === x.high && num.low > x.low)))) {
			throw go$panic(new Go$String("strconv: num > den<<shift in adjustLastDigitFixed"));
		}
		if ((x$1 = go$mul64(new Go$Uint64(0, 2), nonAsciiName), x$2 = go$shiftLeft64(den, shift), (x$1.high > x$2.high || (x$1.high === x$2.high && x$1.low > x$2.low)))) {
			throw go$panic(new Go$String("strconv: \xCE\xB5 > (den<<shift)/2"));
		}
		if ((x$3 = go$mul64(new Go$Uint64(0, 2), (new Go$Uint64(num.high + nonAsciiName.high, num.low + nonAsciiName.low))), x$4 = go$shiftLeft64(den, shift), (x$3.high < x$4.high || (x$3.high === x$4.high && x$3.low < x$4.low)))) {
			return true;
		}
		if ((x$5 = go$mul64(new Go$Uint64(0, 2), (new Go$Uint64(num.high - nonAsciiName.high, num.low - nonAsciiName.low))), x$6 = go$shiftLeft64(den, shift), (x$5.high > x$6.high || (x$5.high === x$6.high && x$5.low > x$6.low)))) {
			i = d.nd - 1 >> 0;
			while (i >= 0) {
				if ((x$7 = d.d, ((i < 0 || i >= x$7.length) ? go$throwRuntimeError("index out of range") : x$7.array[x$7.offset + i])) === 57) {
					d.nd = d.nd - 1 >> 0;
				} else {
					break;
				}
				i = i - 1 >> 0;
			}
			if (i < 0) {
				(x$8 = d.d, (0 < 0 || 0 >= x$8.length) ? go$throwRuntimeError("index out of range") : x$8.array[x$8.offset + 0] = 49);
				d.nd = 1;
				d.dp = d.dp + 1 >> 0;
			} else {
				_lhs = d.d; _index = i; (_index < 0 || _index >= _lhs.length) ? go$throwRuntimeError("index out of range") : _lhs.array[_lhs.offset + _index] = ((_index < 0 || _index >= _lhs.length) ? go$throwRuntimeError("index out of range") : _lhs.array[_lhs.offset + _index]) + 1 << 24 >>> 24;
			}
			return true;
		}
		return false;
	};
	extFloat.Ptr.prototype.ShortestDecimal = function(d, lower, upper) {
		var f, x, x$1, y, x$2, y$1, buf, n, v, v1, x$3, nd, i, x$4, _tmp, _tmp$1, x$5, x$6, exp10, x$7, x$8, shift, integer, x$9, x$10, fraction, x$11, x$12, allowance, x$13, x$14, targetDiff, integerDigits, _tmp$2, _tmp$3, i$1, pow, x$15, i$2, pow$1, _q, digit, x$16, x$17, x$18, currentDiff, digit$1, multiplier, x$19, x$20, x$21, x$22;
		f = this;
		if ((x = f.mant, (x.high === 0 && x.low === 0))) {
			d.nd = 0;
			d.dp = 0;
			d.neg = f.neg;
			return true;
		}
		if ((f.exp === 0) && (x$1 = lower, y = f, (x$1.mant.high === y.mant.high && x$1.mant.low === y.mant.low) && x$1.exp === y.exp && x$1.neg === y.neg) && (x$2 = lower, y$1 = upper, (x$2.mant.high === y$1.mant.high && x$2.mant.low === y$1.mant.low) && x$2.exp === y$1.exp && x$2.neg === y$1.neg)) {
			buf = go$makeNativeArray("Uint8", 24, function() { return 0; });
			n = 23;
			v = f.mant;
			while ((v.high > 0 || (v.high === 0 && v.low > 0))) {
				v1 = go$div64(v, new Go$Uint64(0, 10), false);
				v = (x$3 = go$mul64(new Go$Uint64(0, 10), v1), new Go$Uint64(v.high - x$3.high, v.low - x$3.low));
				buf[n] = (new Go$Uint64(v.high + 0, v.low + 48).low << 24 >>> 24);
				n = n - 1 >> 0;
				v = v1;
			}
			nd = (24 - n >> 0) - 1 >> 0;
			i = 0;
			while (i < nd) {
				(x$4 = d.d, (i < 0 || i >= x$4.length) ? go$throwRuntimeError("index out of range") : x$4.array[x$4.offset + i] = buf[((n + 1 >> 0) + i >> 0)]);
				i = i + 1 >> 0;
			}
			_tmp = nd; _tmp$1 = nd; d.nd = _tmp; d.dp = _tmp$1;
			while (d.nd > 0 && ((x$5 = d.d, x$6 = d.nd - 1 >> 0, ((x$6 < 0 || x$6 >= x$5.length) ? go$throwRuntimeError("index out of range") : x$5.array[x$5.offset + x$6])) === 48)) {
				d.nd = d.nd - 1 >> 0;
			}
			if (d.nd === 0) {
				d.dp = 0;
			}
			d.neg = f.neg;
			return true;
		}
		upper.Normalize();
		if (f.exp > upper.exp) {
			f.mant = go$shiftLeft64(f.mant, (((f.exp - upper.exp >> 0) >>> 0)));
			f.exp = upper.exp;
		}
		if (lower.exp > upper.exp) {
			lower.mant = go$shiftLeft64(lower.mant, (((lower.exp - upper.exp >> 0) >>> 0)));
			lower.exp = upper.exp;
		}
		exp10 = frexp10Many(lower, f, upper);
		upper.mant = (x$7 = upper.mant, new Go$Uint64(x$7.high + 0, x$7.low + 1));
		lower.mant = (x$8 = lower.mant, new Go$Uint64(x$8.high - 0, x$8.low - 1));
		shift = (-upper.exp >>> 0);
		integer = (go$shiftRightUint64(upper.mant, shift).low >>> 0);
		fraction = (x$9 = upper.mant, x$10 = go$shiftLeft64(new Go$Uint64(0, integer), shift), new Go$Uint64(x$9.high - x$10.high, x$9.low - x$10.low));
		allowance = (x$11 = upper.mant, x$12 = lower.mant, new Go$Uint64(x$11.high - x$12.high, x$11.low - x$12.low));
		targetDiff = (x$13 = upper.mant, x$14 = f.mant, new Go$Uint64(x$13.high - x$14.high, x$13.low - x$14.low));
		integerDigits = 0;
		_tmp$2 = 0; _tmp$3 = new Go$Uint64(0, 1); i$1 = _tmp$2; pow = _tmp$3;
		while (i$1 < 20) {
			if ((x$15 = new Go$Uint64(0, integer), (pow.high > x$15.high || (pow.high === x$15.high && pow.low > x$15.low)))) {
				integerDigits = i$1;
				break;
			}
			pow = go$mul64(pow, new Go$Uint64(0, 10));
			i$1 = i$1 + 1 >> 0;
		}
		i$2 = 0;
		while (i$2 < integerDigits) {
			pow$1 = uint64pow10[((integerDigits - i$2 >> 0) - 1 >> 0)];
			digit = (_q = integer / (pow$1.low >>> 0), (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >>> 0 : go$throwRuntimeError("integer divide by zero"));
			(x$16 = d.d, (i$2 < 0 || i$2 >= x$16.length) ? go$throwRuntimeError("index out of range") : x$16.array[x$16.offset + i$2] = ((digit + 48 >>> 0) << 24 >>> 24));
			integer = integer - ((x$17 = (pow$1.low >>> 0), (((digit >>> 16 << 16) * x$17 >>> 0) + (digit << 16 >>> 16) * x$17) >>> 0)) >>> 0;
			currentDiff = (x$18 = go$shiftLeft64(new Go$Uint64(0, integer), shift), new Go$Uint64(x$18.high + fraction.high, x$18.low + fraction.low));
			if ((currentDiff.high < allowance.high || (currentDiff.high === allowance.high && currentDiff.low < allowance.low))) {
				d.nd = i$2 + 1 >> 0;
				d.dp = integerDigits + exp10 >> 0;
				d.neg = f.neg;
				return adjustLastDigit(d, currentDiff, targetDiff, allowance, go$shiftLeft64(pow$1, shift), new Go$Uint64(0, 2));
			}
			i$2 = i$2 + 1 >> 0;
		}
		d.nd = integerDigits;
		d.dp = d.nd + exp10 >> 0;
		d.neg = f.neg;
		digit$1 = 0;
		multiplier = new Go$Uint64(0, 1);
		while (true) {
			fraction = go$mul64(fraction, new Go$Uint64(0, 10));
			multiplier = go$mul64(multiplier, new Go$Uint64(0, 10));
			digit$1 = (go$shiftRightUint64(fraction, shift).low >> 0);
			(x$19 = d.d, x$20 = d.nd, (x$20 < 0 || x$20 >= x$19.length) ? go$throwRuntimeError("index out of range") : x$19.array[x$19.offset + x$20] = ((digit$1 + 48 >> 0) << 24 >>> 24));
			d.nd = d.nd + 1 >> 0;
			fraction = (x$21 = go$shiftLeft64(new Go$Uint64(0, digit$1), shift), new Go$Uint64(fraction.high - x$21.high, fraction.low - x$21.low));
			if ((x$22 = go$mul64(allowance, multiplier), (fraction.high < x$22.high || (fraction.high === x$22.high && fraction.low < x$22.low)))) {
				return adjustLastDigit(d, fraction, go$mul64(targetDiff, multiplier), go$mul64(allowance, multiplier), go$shiftLeft64(new Go$Uint64(0, 1), shift), go$mul64(multiplier, new Go$Uint64(0, 2)));
			}
		}
	};
	extFloat.prototype.ShortestDecimal = function(d, lower, upper) { return this.go$val.ShortestDecimal(d, lower, upper); };
	adjustLastDigit = function(d, currentDiff, targetDiff, maxDiff, ulpDecimal, ulpBinary) {
		var x, x$1, x$2, x$3, _lhs, _index, x$4, x$5, x$6, x$7, x$8, x$9, x$10;
		if ((x = go$mul64(new Go$Uint64(0, 2), ulpBinary), (ulpDecimal.high < x.high || (ulpDecimal.high === x.high && ulpDecimal.low < x.low)))) {
			return false;
		}
		while ((x$1 = (x$2 = (x$3 = go$div64(ulpDecimal, new Go$Uint64(0, 2), false), new Go$Uint64(currentDiff.high + x$3.high, currentDiff.low + x$3.low)), new Go$Uint64(x$2.high + ulpBinary.high, x$2.low + ulpBinary.low)), (x$1.high < targetDiff.high || (x$1.high === targetDiff.high && x$1.low < targetDiff.low)))) {
			_lhs = d.d; _index = d.nd - 1 >> 0; (_index < 0 || _index >= _lhs.length) ? go$throwRuntimeError("index out of range") : _lhs.array[_lhs.offset + _index] = ((_index < 0 || _index >= _lhs.length) ? go$throwRuntimeError("index out of range") : _lhs.array[_lhs.offset + _index]) - 1 << 24 >>> 24;
			currentDiff = (x$4 = ulpDecimal, new Go$Uint64(currentDiff.high + x$4.high, currentDiff.low + x$4.low));
		}
		if ((x$5 = new Go$Uint64(currentDiff.high + ulpDecimal.high, currentDiff.low + ulpDecimal.low), x$6 = (x$7 = (x$8 = go$div64(ulpDecimal, new Go$Uint64(0, 2), false), new Go$Uint64(targetDiff.high + x$8.high, targetDiff.low + x$8.low)), new Go$Uint64(x$7.high + ulpBinary.high, x$7.low + ulpBinary.low)), (x$5.high < x$6.high || (x$5.high === x$6.high && x$5.low <= x$6.low)))) {
			return false;
		}
		if ((currentDiff.high < ulpBinary.high || (currentDiff.high === ulpBinary.high && currentDiff.low < ulpBinary.low)) || (x$9 = new Go$Uint64(maxDiff.high - ulpBinary.high, maxDiff.low - ulpBinary.low), (currentDiff.high > x$9.high || (currentDiff.high === x$9.high && currentDiff.low > x$9.low)))) {
			return false;
		}
		if ((d.nd === 1) && ((x$10 = d.d, ((0 < 0 || 0 >= x$10.length) ? go$throwRuntimeError("index out of range") : x$10.array[x$10.offset + 0])) === 48)) {
			d.nd = 0;
			d.dp = 0;
		}
		return true;
	};
	AppendFloat = go$pkg.AppendFloat = function(dst, f, fmt, prec, bitSize) {
		return genericFtoa(dst, f, fmt, prec, bitSize);
	};
	genericFtoa = function(dst, val, fmt, prec, bitSize) {
		var bits, flt, _ref, x, neg, y, exp, x$1, x$2, mant, _ref$1, y$1, s, x$3, digs, ok, shortest, f, _tuple, _struct, lower, _struct$1, upper, buf, _ref$2, digits, _ref$3, buf$1, f$1, _struct$2;
		bits = new Go$Uint64(0, 0);
		flt = (go$ptrType(floatInfo)).nil;
		_ref = bitSize;
		if (_ref === 32) {
			bits = new Go$Uint64(0, math.Float32bits(val));
			flt = float32info;
		} else if (_ref === 64) {
			bits = math.Float64bits(val);
			flt = float64info;
		} else {
			throw go$panic(new Go$String("strconv: illegal AppendFloat/FormatFloat bitSize"));
		}
		neg = !((x = go$shiftRightUint64(bits, ((flt.expbits + flt.mantbits >>> 0))), (x.high === 0 && x.low === 0)));
		exp = (go$shiftRightUint64(bits, flt.mantbits).low >> 0) & ((((y = flt.expbits, y < 32 ? (1 << y) : 0) >> 0) - 1 >> 0));
		mant = (x$1 = (x$2 = go$shiftLeft64(new Go$Uint64(0, 1), flt.mantbits), new Go$Uint64(x$2.high - 0, x$2.low - 1)), new Go$Uint64(bits.high & x$1.high, (bits.low & x$1.low) >>> 0));
		_ref$1 = exp;
		if (_ref$1 === (((y$1 = flt.expbits, y$1 < 32 ? (1 << y$1) : 0) >> 0) - 1 >> 0)) {
			s = "";
			if (!((mant.high === 0 && mant.low === 0))) {
				s = "NaN";
			} else if (neg) {
				s = "-Inf";
			} else {
				s = "+Inf";
			}
			return go$appendSlice(dst, new (go$sliceType(Go$Uint8))(go$stringToBytes(s)));
		} else if (_ref$1 === 0) {
			exp = exp + 1 >> 0;
		} else {
			mant = (x$3 = go$shiftLeft64(new Go$Uint64(0, 1), flt.mantbits), new Go$Uint64(mant.high | x$3.high, (mant.low | x$3.low) >>> 0));
		}
		exp = exp + (flt.bias) >> 0;
		if (fmt === 98) {
			return fmtB(dst, neg, mant, exp, flt);
		}
		if (!optimize) {
			return bigFtoa(dst, prec, fmt, neg, mant, exp, flt);
		}
		digs = new decimalSlice.Ptr();
		ok = false;
		shortest = prec < 0;
		if (shortest) {
			f = new extFloat.Ptr();
			_tuple = f.AssignComputeBounds(mant, exp, neg, flt); lower = (_struct = _tuple[0], new extFloat.Ptr(_struct.mant, _struct.exp, _struct.neg)); upper = (_struct$1 = _tuple[1], new extFloat.Ptr(_struct$1.mant, _struct$1.exp, _struct$1.neg));
			buf = go$makeNativeArray("Uint8", 32, function() { return 0; });
			digs.d = new (go$sliceType(Go$Uint8))(buf);
			ok = f.ShortestDecimal(digs, lower, upper);
			if (!ok) {
				return bigFtoa(dst, prec, fmt, neg, mant, exp, flt);
			}
			_ref$2 = fmt;
			if (_ref$2 === 101 || _ref$2 === 69) {
				prec = digs.nd - 1 >> 0;
			} else if (_ref$2 === 102) {
				prec = max(digs.nd - digs.dp >> 0, 0);
			} else if (_ref$2 === 103 || _ref$2 === 71) {
				prec = digs.nd;
			}
		} else if (!((fmt === 102))) {
			digits = prec;
			_ref$3 = fmt;
			if (_ref$3 === 101 || _ref$3 === 69) {
				digits = digits + 1 >> 0;
			} else if (_ref$3 === 103 || _ref$3 === 71) {
				if (prec === 0) {
					prec = 1;
				}
				digits = prec;
			}
			if (digits <= 15) {
				buf$1 = go$makeNativeArray("Uint8", 24, function() { return 0; });
				digs.d = new (go$sliceType(Go$Uint8))(buf$1);
				f$1 = new extFloat.Ptr(mant, exp - (flt.mantbits >> 0) >> 0, neg);
				ok = f$1.FixedDecimal(digs, digits);
			}
		}
		if (!ok) {
			return bigFtoa(dst, prec, fmt, neg, mant, exp, flt);
		}
		return formatDigits(dst, shortest, neg, (_struct$2 = digs, new decimalSlice.Ptr(_struct$2.d, _struct$2.nd, _struct$2.dp, _struct$2.neg)), prec, fmt);
	};
	bigFtoa = function(dst, prec, fmt, neg, mant, exp, flt) {
		var d, digs, shortest, _ref, _ref$1, _struct;
		d = new decimal.Ptr();
		d.Assign(mant);
		d.Shift(exp - (flt.mantbits >> 0) >> 0);
		digs = new decimalSlice.Ptr();
		shortest = prec < 0;
		if (shortest) {
			roundShortest(d, mant, exp, flt);
			digs = new decimalSlice.Ptr(new (go$sliceType(Go$Uint8))(d.d), d.nd, d.dp, false);
			_ref = fmt;
			if (_ref === 101 || _ref === 69) {
				prec = digs.nd - 1 >> 0;
			} else if (_ref === 102) {
				prec = max(digs.nd - digs.dp >> 0, 0);
			} else if (_ref === 103 || _ref === 71) {
				prec = digs.nd;
			}
		} else {
			_ref$1 = fmt;
			if (_ref$1 === 101 || _ref$1 === 69) {
				d.Round(prec + 1 >> 0);
			} else if (_ref$1 === 102) {
				d.Round(d.dp + prec >> 0);
			} else if (_ref$1 === 103 || _ref$1 === 71) {
				if (prec === 0) {
					prec = 1;
				}
				d.Round(prec);
			}
			digs = new decimalSlice.Ptr(new (go$sliceType(Go$Uint8))(d.d), d.nd, d.dp, false);
		}
		return formatDigits(dst, shortest, neg, (_struct = digs, new decimalSlice.Ptr(_struct.d, _struct.nd, _struct.dp, _struct.neg)), prec, fmt);
	};
	formatDigits = function(dst, shortest, neg, digs, prec, fmt) {
		var _ref, _struct, _struct$1, eprec, exp, _struct$2, _struct$3;
		_ref = fmt;
		if (_ref === 101 || _ref === 69) {
			return fmtE(dst, neg, (_struct = digs, new decimalSlice.Ptr(_struct.d, _struct.nd, _struct.dp, _struct.neg)), prec, fmt);
		} else if (_ref === 102) {
			return fmtF(dst, neg, (_struct$1 = digs, new decimalSlice.Ptr(_struct$1.d, _struct$1.nd, _struct$1.dp, _struct$1.neg)), prec);
		} else if (_ref === 103 || _ref === 71) {
			eprec = prec;
			if (eprec > digs.nd && digs.nd >= digs.dp) {
				eprec = digs.nd;
			}
			if (shortest) {
				eprec = 6;
			}
			exp = digs.dp - 1 >> 0;
			if (exp < -4 || exp >= eprec) {
				if (prec > digs.nd) {
					prec = digs.nd;
				}
				return fmtE(dst, neg, (_struct$2 = digs, new decimalSlice.Ptr(_struct$2.d, _struct$2.nd, _struct$2.dp, _struct$2.neg)), prec - 1 >> 0, (fmt + 101 << 24 >>> 24) - 103 << 24 >>> 24);
			}
			if (prec > digs.dp) {
				prec = digs.nd;
			}
			return fmtF(dst, neg, (_struct$3 = digs, new decimalSlice.Ptr(_struct$3.d, _struct$3.nd, _struct$3.dp, _struct$3.neg)), max(prec - digs.dp >> 0, 0));
		}
		return go$append(dst, 37, fmt);
	};
	roundShortest = function(d, mant, exp, flt) {
		var minexp, x, x$1, upper, x$2, mantlo, explo, x$3, x$4, lower, x$5, x$6, inclusive, i, _tmp, _tmp$1, _tmp$2, l, m, u, okdown, okup;
		if ((mant.high === 0 && mant.low === 0)) {
			d.nd = 0;
			return;
		}
		minexp = flt.bias + 1 >> 0;
		if (exp > minexp && (x = (d.dp - d.nd >> 0), (((332 >>> 16 << 16) * x >> 0) + (332 << 16 >>> 16) * x) >> 0) >= (x$1 = (exp - (flt.mantbits >> 0) >> 0), (((100 >>> 16 << 16) * x$1 >> 0) + (100 << 16 >>> 16) * x$1) >> 0)) {
			return;
		}
		upper = new decimal.Ptr();
		upper.Assign((x$2 = go$mul64(mant, new Go$Uint64(0, 2)), new Go$Uint64(x$2.high + 0, x$2.low + 1)));
		upper.Shift((exp - (flt.mantbits >> 0) >> 0) - 1 >> 0);
		mantlo = new Go$Uint64(0, 0);
		explo = 0;
		if ((x$3 = go$shiftLeft64(new Go$Uint64(0, 1), flt.mantbits), (mant.high > x$3.high || (mant.high === x$3.high && mant.low > x$3.low))) || (exp === minexp)) {
			mantlo = new Go$Uint64(mant.high - 0, mant.low - 1);
			explo = exp;
		} else {
			mantlo = (x$4 = go$mul64(mant, new Go$Uint64(0, 2)), new Go$Uint64(x$4.high - 0, x$4.low - 1));
			explo = exp - 1 >> 0;
		}
		lower = new decimal.Ptr();
		lower.Assign((x$5 = go$mul64(mantlo, new Go$Uint64(0, 2)), new Go$Uint64(x$5.high + 0, x$5.low + 1)));
		lower.Shift((explo - (flt.mantbits >> 0) >> 0) - 1 >> 0);
		inclusive = (x$6 = go$div64(mant, new Go$Uint64(0, 2), true), (x$6.high === 0 && x$6.low === 0));
		i = 0;
		while (i < d.nd) {
			_tmp = 0; _tmp$1 = 0; _tmp$2 = 0; l = _tmp; m = _tmp$1; u = _tmp$2;
			if (i < lower.nd) {
				l = lower.d[i];
			} else {
				l = 48;
			}
			m = d.d[i];
			if (i < upper.nd) {
				u = upper.d[i];
			} else {
				u = 48;
			}
			okdown = !((l === m)) || (inclusive && (l === m) && ((i + 1 >> 0) === lower.nd));
			okup = !((m === u)) && (inclusive || (m + 1 << 24 >>> 24) < u || (i + 1 >> 0) < upper.nd);
			if (okdown && okup) {
				d.Round(i + 1 >> 0);
				return;
			} else if (okdown) {
				d.RoundDown(i + 1 >> 0);
				return;
			} else if (okup) {
				d.RoundUp(i + 1 >> 0);
				return;
			}
			i = i + 1 >> 0;
		}
	};
	fmtE = function(dst, neg, d, prec, fmt) {
		var ch, x, i, m, x$1, exp, buf, i$1, _r, _q, _ref;
		if (neg) {
			dst = go$append(dst, 45);
		}
		ch = 48;
		if (!((d.nd === 0))) {
			ch = (x = d.d, ((0 < 0 || 0 >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + 0]));
		}
		dst = go$append(dst, ch);
		if (prec > 0) {
			dst = go$append(dst, 46);
			i = 1;
			m = ((d.nd + prec >> 0) + 1 >> 0) - max(d.nd, prec + 1 >> 0) >> 0;
			while (i < m) {
				dst = go$append(dst, (x$1 = d.d, ((i < 0 || i >= x$1.length) ? go$throwRuntimeError("index out of range") : x$1.array[x$1.offset + i])));
				i = i + 1 >> 0;
			}
			while (i <= prec) {
				dst = go$append(dst, 48);
				i = i + 1 >> 0;
			}
		}
		dst = go$append(dst, fmt);
		exp = d.dp - 1 >> 0;
		if (d.nd === 0) {
			exp = 0;
		}
		if (exp < 0) {
			ch = 45;
			exp = -exp;
		} else {
			ch = 43;
		}
		dst = go$append(dst, ch);
		buf = go$makeNativeArray("Uint8", 3, function() { return 0; });
		i$1 = 3;
		while (exp >= 10) {
			i$1 = i$1 - 1 >> 0;
			buf[i$1] = (((_r = exp % 10, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) + 48 >> 0) << 24 >>> 24);
			exp = (_q = exp / 10, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		}
		i$1 = i$1 - 1 >> 0;
		buf[i$1] = ((exp + 48 >> 0) << 24 >>> 24);
		_ref = i$1;
		if (_ref === 0) {
			dst = go$append(dst, buf[0], buf[1], buf[2]);
		} else if (_ref === 1) {
			dst = go$append(dst, buf[1], buf[2]);
		} else if (_ref === 2) {
			dst = go$append(dst, 48, buf[2]);
		}
		return dst;
	};
	fmtF = function(dst, neg, d, prec) {
		var i, x, i$1, ch, j, x$1;
		if (neg) {
			dst = go$append(dst, 45);
		}
		if (d.dp > 0) {
			i = 0;
			i = 0;
			while (i < d.dp && i < d.nd) {
				dst = go$append(dst, (x = d.d, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i])));
				i = i + 1 >> 0;
			}
			while (i < d.dp) {
				dst = go$append(dst, 48);
				i = i + 1 >> 0;
			}
		} else {
			dst = go$append(dst, 48);
		}
		if (prec > 0) {
			dst = go$append(dst, 46);
			i$1 = 0;
			while (i$1 < prec) {
				ch = 48;
				j = d.dp + i$1 >> 0;
				if (0 <= j && j < d.nd) {
					ch = (x$1 = d.d, ((j < 0 || j >= x$1.length) ? go$throwRuntimeError("index out of range") : x$1.array[x$1.offset + j]));
				}
				dst = go$append(dst, ch);
				i$1 = i$1 + 1 >> 0;
			}
		}
		return dst;
	};
	fmtB = function(dst, neg, mant, exp, flt) {
		var buf, w, esign, n, _r, _q, x;
		buf = go$makeNativeArray("Uint8", 50, function() { return 0; });
		w = 50;
		exp = exp - ((flt.mantbits >> 0)) >> 0;
		esign = 43;
		if (exp < 0) {
			esign = 45;
			exp = -exp;
		}
		n = 0;
		while (exp > 0 || n < 1) {
			n = n + 1 >> 0;
			w = w - 1 >> 0;
			buf[w] = (((_r = exp % 10, _r === _r ? _r : go$throwRuntimeError("integer divide by zero")) + 48 >> 0) << 24 >>> 24);
			exp = (_q = exp / 10, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero"));
		}
		w = w - 1 >> 0;
		buf[w] = esign;
		w = w - 1 >> 0;
		buf[w] = 112;
		n = 0;
		while ((mant.high > 0 || (mant.high === 0 && mant.low > 0)) || n < 1) {
			n = n + 1 >> 0;
			w = w - 1 >> 0;
			buf[w] = ((x = go$div64(mant, new Go$Uint64(0, 10), true), new Go$Uint64(x.high + 0, x.low + 48)).low << 24 >>> 24);
			mant = go$div64(mant, new Go$Uint64(0, 10), false);
		}
		if (neg) {
			w = w - 1 >> 0;
			buf[w] = 45;
		}
		return go$appendSlice(dst, go$subslice(new (go$sliceType(Go$Uint8))(buf), w));
	};
	max = function(a, b) {
		if (a > b) {
			return a;
		}
		return b;
	};
	FormatInt = go$pkg.FormatInt = function(i, base) {
		var _tuple, s;
		_tuple = formatBits((go$sliceType(Go$Uint8)).nil, new Go$Uint64(i.high, i.low), base, (i.high < 0 || (i.high === 0 && i.low < 0)), false); s = _tuple[1];
		return s;
	};
	Itoa = go$pkg.Itoa = function(i) {
		return FormatInt(new Go$Int64(0, i), 10);
	};
	formatBits = function(dst, u, base, neg, append_) {
		var d, s, a, i, q, x, j, q$1, x$1, s$1, b, m, b$1;
		d = (go$sliceType(Go$Uint8)).nil;
		s = "";
		if (base < 2 || base > 36) {
			throw go$panic(new Go$String("strconv: illegal AppendInt/FormatInt base"));
		}
		a = go$makeNativeArray("Uint8", 65, function() { return 0; });
		i = 65;
		if (neg) {
			u = new Go$Uint64(-u.high, -u.low);
		}
		if (base === 10) {
			while ((u.high > 0 || (u.high === 0 && u.low >= 100))) {
				i = i - 2 >> 0;
				q = go$div64(u, new Go$Uint64(0, 100), false);
				j = ((x = go$mul64(q, new Go$Uint64(0, 100)), new Go$Uint64(u.high - x.high, u.low - x.low)).low >>> 0);
				a[(i + 1 >> 0)] = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789".charCodeAt(j);
				a[(i + 0 >> 0)] = "0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999".charCodeAt(j);
				u = q;
			}
			if ((u.high > 0 || (u.high === 0 && u.low >= 10))) {
				i = i - 1 >> 0;
				q$1 = go$div64(u, new Go$Uint64(0, 10), false);
				a[i] = "0123456789abcdefghijklmnopqrstuvwxyz".charCodeAt(((x$1 = go$mul64(q$1, new Go$Uint64(0, 10)), new Go$Uint64(u.high - x$1.high, u.low - x$1.low)).low >>> 0));
				u = q$1;
			}
		} else {
			s$1 = shifts[base];
			if (s$1 > 0) {
				b = new Go$Uint64(0, base);
				m = (b.low >>> 0) - 1 >>> 0;
				while ((u.high > b.high || (u.high === b.high && u.low >= b.low))) {
					i = i - 1 >> 0;
					a[i] = "0123456789abcdefghijklmnopqrstuvwxyz".charCodeAt((((u.low >>> 0) & m) >>> 0));
					u = go$shiftRightUint64(u, (s$1));
				}
			} else {
				b$1 = new Go$Uint64(0, base);
				while ((u.high > b$1.high || (u.high === b$1.high && u.low >= b$1.low))) {
					i = i - 1 >> 0;
					a[i] = "0123456789abcdefghijklmnopqrstuvwxyz".charCodeAt((go$div64(u, b$1, true).low >>> 0));
					u = go$div64(u, (b$1), false);
				}
			}
		}
		i = i - 1 >> 0;
		a[i] = "0123456789abcdefghijklmnopqrstuvwxyz".charCodeAt((u.low >>> 0));
		if (neg) {
			i = i - 1 >> 0;
			a[i] = 45;
		}
		if (append_) {
			d = go$appendSlice(dst, go$subslice(new (go$sliceType(Go$Uint8))(a), i));
			return [d, s];
		}
		s = go$bytesToString(go$subslice(new (go$sliceType(Go$Uint8))(a), i));
		return [d, s];
	};
	quoteWith = function(s, quote, ASCIIonly) {
		var runeTmp, _q, x, buf, width, r, _tuple, n, _ref, s$1, s$2;
		runeTmp = go$makeNativeArray("Uint8", 4, function() { return 0; });
		buf = (go$sliceType(Go$Uint8)).make(0, (_q = (x = s.length, (((3 >>> 16 << 16) * x >> 0) + (3 << 16 >>> 16) * x) >> 0) / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")), function() { return 0; });
		buf = go$append(buf, quote);
		width = 0;
		while (s.length > 0) {
			r = (s.charCodeAt(0) >> 0);
			width = 1;
			if (r >= 128) {
				_tuple = utf8.DecodeRuneInString(s); r = _tuple[0]; width = _tuple[1];
			}
			if ((width === 1) && (r === 65533)) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\x")));
				buf = go$append(buf, "0123456789abcdef".charCodeAt((s.charCodeAt(0) >>> 4 << 24 >>> 24)));
				buf = go$append(buf, "0123456789abcdef".charCodeAt(((s.charCodeAt(0) & 15) >>> 0)));
				s = s.substring(width);
				continue;
			}
			if ((r === (quote >> 0)) || (r === 92)) {
				buf = go$append(buf, 92);
				buf = go$append(buf, (r << 24 >>> 24));
				s = s.substring(width);
				continue;
			}
			if (ASCIIonly) {
				if (r < 128 && IsPrint(r)) {
					buf = go$append(buf, (r << 24 >>> 24));
					s = s.substring(width);
					continue;
				}
			} else if (IsPrint(r)) {
				n = utf8.EncodeRune(new (go$sliceType(Go$Uint8))(runeTmp), r);
				buf = go$appendSlice(buf, go$subslice(new (go$sliceType(Go$Uint8))(runeTmp), 0, n));
				s = s.substring(width);
				continue;
			}
			_ref = r;
			if (_ref === 7) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\a")));
			} else if (_ref === 8) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\b")));
			} else if (_ref === 12) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\f")));
			} else if (_ref === 10) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\n")));
			} else if (_ref === 13) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\r")));
			} else if (_ref === 9) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\t")));
			} else if (_ref === 11) {
				buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\v")));
			} else {
				if (r < 32) {
					buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\x")));
					buf = go$append(buf, "0123456789abcdef".charCodeAt((s.charCodeAt(0) >>> 4 << 24 >>> 24)));
					buf = go$append(buf, "0123456789abcdef".charCodeAt(((s.charCodeAt(0) & 15) >>> 0)));
				} else if (r > 1114111) {
					r = 65533;
					buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\u")));
					s$1 = 12;
					while (s$1 >= 0) {
						buf = go$append(buf, "0123456789abcdef".charCodeAt((((r >> go$min((s$1 >>> 0), 31)) >> 0) & 15)));
						s$1 = s$1 - 4 >> 0;
					}
				} else if (r < 65536) {
					buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\u")));
					s$1 = 12;
					while (s$1 >= 0) {
						buf = go$append(buf, "0123456789abcdef".charCodeAt((((r >> go$min((s$1 >>> 0), 31)) >> 0) & 15)));
						s$1 = s$1 - 4 >> 0;
					}
				} else {
					buf = go$appendSlice(buf, new (go$sliceType(Go$Uint8))(go$stringToBytes("\\U")));
					s$2 = 28;
					while (s$2 >= 0) {
						buf = go$append(buf, "0123456789abcdef".charCodeAt((((r >> go$min((s$2 >>> 0), 31)) >> 0) & 15)));
						s$2 = s$2 - 4 >> 0;
					}
				}
			}
			s = s.substring(width);
		}
		buf = go$append(buf, quote);
		return go$bytesToString(buf);
	};
	Quote = go$pkg.Quote = function(s) {
		return quoteWith(s, 34, false);
	};
	QuoteToASCII = go$pkg.QuoteToASCII = function(s) {
		return quoteWith(s, 34, true);
	};
	QuoteRune = go$pkg.QuoteRune = function(r) {
		return quoteWith(go$encodeRune(r), 39, false);
	};
	AppendQuoteRune = go$pkg.AppendQuoteRune = function(dst, r) {
		return go$appendSlice(dst, new (go$sliceType(Go$Uint8))(go$stringToBytes(QuoteRune(r))));
	};
	QuoteRuneToASCII = go$pkg.QuoteRuneToASCII = function(r) {
		return quoteWith(go$encodeRune(r), 39, true);
	};
	AppendQuoteRuneToASCII = go$pkg.AppendQuoteRuneToASCII = function(dst, r) {
		return go$appendSlice(dst, new (go$sliceType(Go$Uint8))(go$stringToBytes(QuoteRuneToASCII(r))));
	};
	CanBackquote = go$pkg.CanBackquote = function(s) {
		var i;
		i = 0;
		while (i < s.length) {
			if ((s.charCodeAt(i) < 32 && !((s.charCodeAt(i) === 9))) || (s.charCodeAt(i) === 96)) {
				return false;
			}
			i = i + 1 >> 0;
		}
		return true;
	};
	unhex = function(b) {
		var v, ok, c, _tmp, _tmp$1, _tmp$2, _tmp$3, _tmp$4, _tmp$5;
		v = 0;
		ok = false;
		c = (b >> 0);
		if (48 <= c && c <= 57) {
			_tmp = c - 48 >> 0; _tmp$1 = true; v = _tmp; ok = _tmp$1;
			return [v, ok];
		} else if (97 <= c && c <= 102) {
			_tmp$2 = (c - 97 >> 0) + 10 >> 0; _tmp$3 = true; v = _tmp$2; ok = _tmp$3;
			return [v, ok];
		} else if (65 <= c && c <= 70) {
			_tmp$4 = (c - 65 >> 0) + 10 >> 0; _tmp$5 = true; v = _tmp$4; ok = _tmp$5;
			return [v, ok];
		}
		return [v, ok];
	};
	UnquoteChar = go$pkg.UnquoteChar = function(s, quote) {
		var value, multibyte, tail, err, c, _tuple, r, size, _tmp, _tmp$1, _tmp$2, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, c$1, _ref, n, _ref$1, v, j, _tuple$1, x, ok, v$1, j$1, x$1;
		value = 0;
		multibyte = false;
		tail = "";
		err = null;
		c = s.charCodeAt(0);
		if ((c === quote) && ((quote === 39) || (quote === 34))) {
			err = go$pkg.ErrSyntax;
			return [value, multibyte, tail, err];
		} else if (c >= 128) {
			_tuple = utf8.DecodeRuneInString(s); r = _tuple[0]; size = _tuple[1];
			_tmp = r; _tmp$1 = true; _tmp$2 = s.substring(size); _tmp$3 = null; value = _tmp; multibyte = _tmp$1; tail = _tmp$2; err = _tmp$3;
			return [value, multibyte, tail, err];
		} else if (!((c === 92))) {
			_tmp$4 = (s.charCodeAt(0) >> 0); _tmp$5 = false; _tmp$6 = s.substring(1); _tmp$7 = null; value = _tmp$4; multibyte = _tmp$5; tail = _tmp$6; err = _tmp$7;
			return [value, multibyte, tail, err];
		}
		if (s.length <= 1) {
			err = go$pkg.ErrSyntax;
			return [value, multibyte, tail, err];
		}
		c$1 = s.charCodeAt(1);
		s = s.substring(2);
		_ref = c$1;
		switch (0) { default: if (_ref === 97) {
			value = 7;
		} else if (_ref === 98) {
			value = 8;
		} else if (_ref === 102) {
			value = 12;
		} else if (_ref === 110) {
			value = 10;
		} else if (_ref === 114) {
			value = 13;
		} else if (_ref === 116) {
			value = 9;
		} else if (_ref === 118) {
			value = 11;
		} else if (_ref === 120 || _ref === 117 || _ref === 85) {
			n = 0;
			_ref$1 = c$1;
			if (_ref$1 === 120) {
				n = 2;
			} else if (_ref$1 === 117) {
				n = 4;
			} else if (_ref$1 === 85) {
				n = 8;
			}
			v = 0;
			if (s.length < n) {
				err = go$pkg.ErrSyntax;
				return [value, multibyte, tail, err];
			}
			j = 0;
			while (j < n) {
				_tuple$1 = unhex(s.charCodeAt(j)); x = _tuple$1[0]; ok = _tuple$1[1];
				if (!ok) {
					err = go$pkg.ErrSyntax;
					return [value, multibyte, tail, err];
				}
				v = (v << 4 >> 0) | x;
				j = j + 1 >> 0;
			}
			s = s.substring(n);
			if (c$1 === 120) {
				value = v;
				break;
			}
			if (v > 1114111) {
				err = go$pkg.ErrSyntax;
				return [value, multibyte, tail, err];
			}
			value = v;
			multibyte = true;
		} else if (_ref === 48 || _ref === 49 || _ref === 50 || _ref === 51 || _ref === 52 || _ref === 53 || _ref === 54 || _ref === 55) {
			v$1 = (c$1 >> 0) - 48 >> 0;
			if (s.length < 2) {
				err = go$pkg.ErrSyntax;
				return [value, multibyte, tail, err];
			}
			j$1 = 0;
			while (j$1 < 2) {
				x$1 = (s.charCodeAt(j$1) >> 0) - 48 >> 0;
				if (x$1 < 0 || x$1 > 7) {
					err = go$pkg.ErrSyntax;
					return [value, multibyte, tail, err];
				}
				v$1 = ((v$1 << 3 >> 0)) | x$1;
				j$1 = j$1 + 1 >> 0;
			}
			s = s.substring(2);
			if (v$1 > 255) {
				err = go$pkg.ErrSyntax;
				return [value, multibyte, tail, err];
			}
			value = v$1;
		} else if (_ref === 92) {
			value = 92;
		} else if (_ref === 39 || _ref === 34) {
			if (!((c$1 === quote))) {
				err = go$pkg.ErrSyntax;
				return [value, multibyte, tail, err];
			}
			value = (c$1 >> 0);
		} else {
			err = go$pkg.ErrSyntax;
			return [value, multibyte, tail, err];
		} }
		tail = s;
		return [value, multibyte, tail, err];
	};
	Unquote = go$pkg.Unquote = function(s) {
		var t, err, n, _tmp, _tmp$1, quote, _tmp$2, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, _tmp$8, _tmp$9, _tmp$10, _tmp$11, _ref, _tmp$12, _tmp$13, _tuple, r, size, _tmp$14, _tmp$15, runeTmp, _q, x, buf, _tuple$1, c, multibyte, ss, err$1, _tmp$16, _tmp$17, n$1, _tmp$18, _tmp$19, _tmp$20, _tmp$21;
		t = "";
		err = null;
		n = s.length;
		if (n < 2) {
			_tmp = ""; _tmp$1 = go$pkg.ErrSyntax; t = _tmp; err = _tmp$1;
			return [t, err];
		}
		quote = s.charCodeAt(0);
		if (!((quote === s.charCodeAt((n - 1 >> 0))))) {
			_tmp$2 = ""; _tmp$3 = go$pkg.ErrSyntax; t = _tmp$2; err = _tmp$3;
			return [t, err];
		}
		s = s.substring(1, (n - 1 >> 0));
		if (quote === 96) {
			if (contains(s, 96)) {
				_tmp$4 = ""; _tmp$5 = go$pkg.ErrSyntax; t = _tmp$4; err = _tmp$5;
				return [t, err];
			}
			_tmp$6 = s; _tmp$7 = null; t = _tmp$6; err = _tmp$7;
			return [t, err];
		}
		if (!((quote === 34)) && !((quote === 39))) {
			_tmp$8 = ""; _tmp$9 = go$pkg.ErrSyntax; t = _tmp$8; err = _tmp$9;
			return [t, err];
		}
		if (contains(s, 10)) {
			_tmp$10 = ""; _tmp$11 = go$pkg.ErrSyntax; t = _tmp$10; err = _tmp$11;
			return [t, err];
		}
		if (!contains(s, 92) && !contains(s, quote)) {
			_ref = quote;
			if (_ref === 34) {
				_tmp$12 = s; _tmp$13 = null; t = _tmp$12; err = _tmp$13;
				return [t, err];
			} else if (_ref === 39) {
				_tuple = utf8.DecodeRuneInString(s); r = _tuple[0]; size = _tuple[1];
				if ((size === s.length) && (!((r === 65533)) || !((size === 1)))) {
					_tmp$14 = s; _tmp$15 = null; t = _tmp$14; err = _tmp$15;
					return [t, err];
				}
			}
		}
		runeTmp = go$makeNativeArray("Uint8", 4, function() { return 0; });
		buf = (go$sliceType(Go$Uint8)).make(0, (_q = (x = s.length, (((3 >>> 16 << 16) * x >> 0) + (3 << 16 >>> 16) * x) >> 0) / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")), function() { return 0; });
		while (s.length > 0) {
			_tuple$1 = UnquoteChar(s, quote); c = _tuple$1[0]; multibyte = _tuple$1[1]; ss = _tuple$1[2]; err$1 = _tuple$1[3];
			if (!(go$interfaceIsEqual(err$1, null))) {
				_tmp$16 = ""; _tmp$17 = err$1; t = _tmp$16; err = _tmp$17;
				return [t, err];
			}
			s = ss;
			if (c < 128 || !multibyte) {
				buf = go$append(buf, (c << 24 >>> 24));
			} else {
				n$1 = utf8.EncodeRune(new (go$sliceType(Go$Uint8))(runeTmp), c);
				buf = go$appendSlice(buf, go$subslice(new (go$sliceType(Go$Uint8))(runeTmp), 0, n$1));
			}
			if ((quote === 39) && !((s.length === 0))) {
				_tmp$18 = ""; _tmp$19 = go$pkg.ErrSyntax; t = _tmp$18; err = _tmp$19;
				return [t, err];
			}
		}
		_tmp$20 = go$bytesToString(buf); _tmp$21 = null; t = _tmp$20; err = _tmp$21;
		return [t, err];
	};
	contains = function(s, c) {
		var i;
		i = 0;
		while (i < s.length) {
			if (s.charCodeAt(i) === c) {
				return true;
			}
			i = i + 1 >> 0;
		}
		return false;
	};
	bsearch16 = function(a, x) {
		var _tmp, _tmp$1, i, j, _q, h;
		_tmp = 0; _tmp$1 = a.length; i = _tmp; j = _tmp$1;
		while (i < j) {
			h = i + (_q = ((j - i >> 0)) / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0;
			if (((h < 0 || h >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + h]) < x) {
				i = h + 1 >> 0;
			} else {
				j = h;
			}
		}
		return i;
	};
	bsearch32 = function(a, x) {
		var _tmp, _tmp$1, i, j, _q, h;
		_tmp = 0; _tmp$1 = a.length; i = _tmp; j = _tmp$1;
		while (i < j) {
			h = i + (_q = ((j - i >> 0)) / 2, (_q === _q && _q !== 1/0 && _q !== -1/0) ? _q >> 0 : go$throwRuntimeError("integer divide by zero")) >> 0;
			if (((h < 0 || h >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + h]) < x) {
				i = h + 1 >> 0;
			} else {
				j = h;
			}
		}
		return i;
	};
	IsPrint = go$pkg.IsPrint = function(r) {
		var _tmp, _tmp$1, _tmp$2, rr, isPrint, isNotPrint, i, x, x$1, j, _tmp$3, _tmp$4, _tmp$5, rr$1, isPrint$1, isNotPrint$1, i$1, x$2, x$3, j$1;
		if (r <= 255) {
			if (32 <= r && r <= 126) {
				return true;
			}
			if (161 <= r && r <= 255) {
				return !((r === 173));
			}
			return false;
		}
		if (0 <= r && r < 65536) {
			_tmp = (r << 16 >>> 16); _tmp$1 = isPrint16; _tmp$2 = isNotPrint16; rr = _tmp; isPrint = _tmp$1; isNotPrint = _tmp$2;
			i = bsearch16(isPrint, rr);
			if (i >= isPrint.length || rr < (x = i & ~1, ((x < 0 || x >= isPrint.length) ? go$throwRuntimeError("index out of range") : isPrint.array[isPrint.offset + x])) || (x$1 = i | 1, ((x$1 < 0 || x$1 >= isPrint.length) ? go$throwRuntimeError("index out of range") : isPrint.array[isPrint.offset + x$1])) < rr) {
				return false;
			}
			j = bsearch16(isNotPrint, rr);
			return j >= isNotPrint.length || !((((j < 0 || j >= isNotPrint.length) ? go$throwRuntimeError("index out of range") : isNotPrint.array[isNotPrint.offset + j]) === rr));
		}
		_tmp$3 = (r >>> 0); _tmp$4 = isPrint32; _tmp$5 = isNotPrint32; rr$1 = _tmp$3; isPrint$1 = _tmp$4; isNotPrint$1 = _tmp$5;
		i$1 = bsearch32(isPrint$1, rr$1);
		if (i$1 >= isPrint$1.length || rr$1 < (x$2 = i$1 & ~1, ((x$2 < 0 || x$2 >= isPrint$1.length) ? go$throwRuntimeError("index out of range") : isPrint$1.array[isPrint$1.offset + x$2])) || (x$3 = i$1 | 1, ((x$3 < 0 || x$3 >= isPrint$1.length) ? go$throwRuntimeError("index out of range") : isPrint$1.array[isPrint$1.offset + x$3])) < rr$1) {
			return false;
		}
		if (r >= 131072) {
			return true;
		}
		r = r - 65536 >> 0;
		j$1 = bsearch16(isNotPrint$1, (r << 16 >>> 16));
		return j$1 >= isNotPrint$1.length || !((((j$1 < 0 || j$1 >= isNotPrint$1.length) ? go$throwRuntimeError("index out of range") : isNotPrint$1.array[isNotPrint$1.offset + j$1]) === (r << 16 >>> 16)));
	};
	go$pkg.init = function() {
		(go$ptrType(decimal)).methods = [["Assign", "", [Go$Uint64], [], false, -1], ["Round", "", [Go$Int], [], false, -1], ["RoundDown", "", [Go$Int], [], false, -1], ["RoundUp", "", [Go$Int], [], false, -1], ["RoundedInteger", "", [], [Go$Uint64], false, -1], ["Shift", "", [Go$Int], [], false, -1], ["String", "", [], [Go$String], false, -1], ["atof32int", "strconv", [], [Go$Float32], false, -1], ["floatBits", "strconv", [(go$ptrType(floatInfo))], [Go$Uint64, Go$Bool], false, -1], ["set", "strconv", [Go$String], [Go$Bool], false, -1]];
		decimal.init([["d", "d", "strconv", (go$arrayType(Go$Uint8, 800)), ""], ["nd", "nd", "strconv", Go$Int, ""], ["dp", "dp", "strconv", Go$Int, ""], ["neg", "neg", "strconv", Go$Bool, ""], ["trunc", "trunc", "strconv", Go$Bool, ""]]);
		leftCheat.init([["delta", "delta", "strconv", Go$Int, ""], ["cutoff", "cutoff", "strconv", Go$String, ""]]);
		(go$ptrType(extFloat)).methods = [["AssignComputeBounds", "", [Go$Uint64, Go$Int, Go$Bool, (go$ptrType(floatInfo))], [extFloat, extFloat], false, -1], ["AssignDecimal", "", [Go$Uint64, Go$Int, Go$Bool, Go$Bool, (go$ptrType(floatInfo))], [Go$Bool], false, -1], ["FixedDecimal", "", [(go$ptrType(decimalSlice)), Go$Int], [Go$Bool], false, -1], ["Multiply", "", [extFloat], [], false, -1], ["Normalize", "", [], [Go$Uint], false, -1], ["ShortestDecimal", "", [(go$ptrType(decimalSlice)), (go$ptrType(extFloat)), (go$ptrType(extFloat))], [Go$Bool], false, -1], ["floatBits", "strconv", [(go$ptrType(floatInfo))], [Go$Uint64, Go$Bool], false, -1], ["frexp10", "strconv", [], [Go$Int, Go$Int], false, -1]];
		extFloat.init([["mant", "mant", "strconv", Go$Uint64, ""], ["exp", "exp", "strconv", Go$Int, ""], ["neg", "neg", "strconv", Go$Bool, ""]]);
		floatInfo.init([["mantbits", "mantbits", "strconv", Go$Uint, ""], ["expbits", "expbits", "strconv", Go$Uint, ""], ["bias", "bias", "strconv", Go$Int, ""]]);
		decimalSlice.init([["d", "d", "strconv", (go$sliceType(Go$Uint8)), ""], ["nd", "nd", "strconv", Go$Int, ""], ["dp", "dp", "strconv", Go$Int, ""], ["neg", "neg", "strconv", Go$Bool, ""]]);
		optimize = true;
		go$pkg.ErrRange = errors.New("value out of range");
		go$pkg.ErrSyntax = errors.New("invalid syntax");
		leftcheats = new (go$sliceType(leftCheat))([new leftCheat.Ptr(0, ""), new leftCheat.Ptr(1, "5"), new leftCheat.Ptr(1, "25"), new leftCheat.Ptr(1, "125"), new leftCheat.Ptr(2, "625"), new leftCheat.Ptr(2, "3125"), new leftCheat.Ptr(2, "15625"), new leftCheat.Ptr(3, "78125"), new leftCheat.Ptr(3, "390625"), new leftCheat.Ptr(3, "1953125"), new leftCheat.Ptr(4, "9765625"), new leftCheat.Ptr(4, "48828125"), new leftCheat.Ptr(4, "244140625"), new leftCheat.Ptr(4, "1220703125"), new leftCheat.Ptr(5, "6103515625"), new leftCheat.Ptr(5, "30517578125"), new leftCheat.Ptr(5, "152587890625"), new leftCheat.Ptr(6, "762939453125"), new leftCheat.Ptr(6, "3814697265625"), new leftCheat.Ptr(6, "19073486328125"), new leftCheat.Ptr(7, "95367431640625"), new leftCheat.Ptr(7, "476837158203125"), new leftCheat.Ptr(7, "2384185791015625"), new leftCheat.Ptr(7, "11920928955078125"), new leftCheat.Ptr(8, "59604644775390625"), new leftCheat.Ptr(8, "298023223876953125"), new leftCheat.Ptr(8, "1490116119384765625"), new leftCheat.Ptr(9, "7450580596923828125")]);
		smallPowersOfTen = go$toNativeArray("Struct", [new extFloat.Ptr(new Go$Uint64(2147483648, 0), -63, false), new extFloat.Ptr(new Go$Uint64(2684354560, 0), -60, false), new extFloat.Ptr(new Go$Uint64(3355443200, 0), -57, false), new extFloat.Ptr(new Go$Uint64(4194304000, 0), -54, false), new extFloat.Ptr(new Go$Uint64(2621440000, 0), -50, false), new extFloat.Ptr(new Go$Uint64(3276800000, 0), -47, false), new extFloat.Ptr(new Go$Uint64(4096000000, 0), -44, false), new extFloat.Ptr(new Go$Uint64(2560000000, 0), -40, false)]);
		powersOfTen = go$toNativeArray("Struct", [new extFloat.Ptr(new Go$Uint64(4203730336, 136053384), -1220, false), new extFloat.Ptr(new Go$Uint64(3132023167, 2722021238), -1193, false), new extFloat.Ptr(new Go$Uint64(2333539104, 810921078), -1166, false), new extFloat.Ptr(new Go$Uint64(3477244234, 1573795306), -1140, false), new extFloat.Ptr(new Go$Uint64(2590748842, 1432697645), -1113, false), new extFloat.Ptr(new Go$Uint64(3860516611, 1025131999), -1087, false), new extFloat.Ptr(new Go$Uint64(2876309015, 3348809418), -1060, false), new extFloat.Ptr(new Go$Uint64(4286034428, 3200048207), -1034, false), new extFloat.Ptr(new Go$Uint64(3193344495, 1097586188), -1007, false), new extFloat.Ptr(new Go$Uint64(2379227053, 2424306748), -980, false), new extFloat.Ptr(new Go$Uint64(3545324584, 827693699), -954, false), new extFloat.Ptr(new Go$Uint64(2641472655, 2913388981), -927, false), new extFloat.Ptr(new Go$Uint64(3936100983, 602835915), -901, false), new extFloat.Ptr(new Go$Uint64(2932623761, 1081627501), -874, false), new extFloat.Ptr(new Go$Uint64(2184974969, 1572261463), -847, false), new extFloat.Ptr(new Go$Uint64(3255866422, 1308317239), -821, false), new extFloat.Ptr(new Go$Uint64(2425809519, 944281679), -794, false), new extFloat.Ptr(new Go$Uint64(3614737867, 629291719), -768, false), new extFloat.Ptr(new Go$Uint64(2693189581, 2545915892), -741, false), new extFloat.Ptr(new Go$Uint64(4013165208, 388672741), -715, false), new extFloat.Ptr(new Go$Uint64(2990041083, 708162190), -688, false), new extFloat.Ptr(new Go$Uint64(2227754207, 3536207675), -661, false), new extFloat.Ptr(new Go$Uint64(3319612455, 450088378), -635, false), new extFloat.Ptr(new Go$Uint64(2473304014, 3139815830), -608, false), new extFloat.Ptr(new Go$Uint64(3685510180, 2103616900), -582, false), new extFloat.Ptr(new Go$Uint64(2745919064, 224385782), -555, false), new extFloat.Ptr(new Go$Uint64(4091738259, 3737383206), -529, false), new extFloat.Ptr(new Go$Uint64(3048582568, 2868871352), -502, false), new extFloat.Ptr(new Go$Uint64(2271371013, 1820084875), -475, false), new extFloat.Ptr(new Go$Uint64(3384606560, 885076051), -449, false), new extFloat.Ptr(new Go$Uint64(2521728396, 2444895829), -422, false), new extFloat.Ptr(new Go$Uint64(3757668132, 1881767613), -396, false), new extFloat.Ptr(new Go$Uint64(2799680927, 3102062735), -369, false), new extFloat.Ptr(new Go$Uint64(4171849679, 2289335700), -343, false), new extFloat.Ptr(new Go$Uint64(3108270227, 2410191823), -316, false), new extFloat.Ptr(new Go$Uint64(2315841784, 3205436779), -289, false), new extFloat.Ptr(new Go$Uint64(3450873173, 1697722806), -263, false), new extFloat.Ptr(new Go$Uint64(2571100870, 3497754540), -236, false), new extFloat.Ptr(new Go$Uint64(3831238852, 707476230), -210, false), new extFloat.Ptr(new Go$Uint64(2854495385, 1769181907), -183, false), new extFloat.Ptr(new Go$Uint64(4253529586, 2197867022), -157, false), new extFloat.Ptr(new Go$Uint64(3169126500, 2450594539), -130, false), new extFloat.Ptr(new Go$Uint64(2361183241, 1867548876), -103, false), new extFloat.Ptr(new Go$Uint64(3518437208, 3793315116), -77, false), new extFloat.Ptr(new Go$Uint64(2621440000, 0), -50, false), new extFloat.Ptr(new Go$Uint64(3906250000, 0), -24, false), new extFloat.Ptr(new Go$Uint64(2910383045, 2892103680), 3, false), new extFloat.Ptr(new Go$Uint64(2168404344, 4170451332), 30, false), new extFloat.Ptr(new Go$Uint64(3231174267, 3372684723), 56, false), new extFloat.Ptr(new Go$Uint64(2407412430, 2078956656), 83, false), new extFloat.Ptr(new Go$Uint64(3587324068, 2884206696), 109, false), new extFloat.Ptr(new Go$Uint64(2672764710, 395977285), 136, false), new extFloat.Ptr(new Go$Uint64(3982729777, 3569679143), 162, false), new extFloat.Ptr(new Go$Uint64(2967364920, 2361961896), 189, false), new extFloat.Ptr(new Go$Uint64(2210859150, 447440347), 216, false), new extFloat.Ptr(new Go$Uint64(3294436857, 1114709402), 242, false), new extFloat.Ptr(new Go$Uint64(2454546732, 2786846552), 269, false), new extFloat.Ptr(new Go$Uint64(3657559652, 443583978), 295, false), new extFloat.Ptr(new Go$Uint64(2725094297, 2599384906), 322, false), new extFloat.Ptr(new Go$Uint64(4060706939, 3028118405), 348, false), new extFloat.Ptr(new Go$Uint64(3025462433, 2044532855), 375, false), new extFloat.Ptr(new Go$Uint64(2254145170, 1536935362), 402, false), new extFloat.Ptr(new Go$Uint64(3358938053, 3365297469), 428, false), new extFloat.Ptr(new Go$Uint64(2502603868, 4204241075), 455, false), new extFloat.Ptr(new Go$Uint64(3729170365, 2577424355), 481, false), new extFloat.Ptr(new Go$Uint64(2778448436, 3677981733), 508, false), new extFloat.Ptr(new Go$Uint64(4140210802, 2744688476), 534, false), new extFloat.Ptr(new Go$Uint64(3084697427, 1424604878), 561, false), new extFloat.Ptr(new Go$Uint64(2298278679, 4062331362), 588, false), new extFloat.Ptr(new Go$Uint64(3424702107, 3546052773), 614, false), new extFloat.Ptr(new Go$Uint64(2551601907, 2065781727), 641, false), new extFloat.Ptr(new Go$Uint64(3802183132, 2535403578), 667, false), new extFloat.Ptr(new Go$Uint64(2832847187, 1558426518), 694, false), new extFloat.Ptr(new Go$Uint64(4221271257, 2762425404), 720, false), new extFloat.Ptr(new Go$Uint64(3145092172, 2812560400), 747, false), new extFloat.Ptr(new Go$Uint64(2343276271, 3057687578), 774, false), new extFloat.Ptr(new Go$Uint64(3491753744, 2790753324), 800, false), new extFloat.Ptr(new Go$Uint64(2601559269, 3918606633), 827, false), new extFloat.Ptr(new Go$Uint64(3876625403, 2711358621), 853, false), new extFloat.Ptr(new Go$Uint64(2888311001, 1648096297), 880, false), new extFloat.Ptr(new Go$Uint64(2151959390, 2057817989), 907, false), new extFloat.Ptr(new Go$Uint64(3206669376, 61660461), 933, false), new extFloat.Ptr(new Go$Uint64(2389154863, 1581580175), 960, false), new extFloat.Ptr(new Go$Uint64(3560118173, 2626467905), 986, false), new extFloat.Ptr(new Go$Uint64(2652494738, 3034782633), 1013, false), new extFloat.Ptr(new Go$Uint64(3952525166, 3135207385), 1039, false), new extFloat.Ptr(new Go$Uint64(2944860731, 2616258155), 1066, false)]);
		uint64pow10 = go$toNativeArray("Uint64", [new Go$Uint64(0, 1), new Go$Uint64(0, 10), new Go$Uint64(0, 100), new Go$Uint64(0, 1000), new Go$Uint64(0, 10000), new Go$Uint64(0, 100000), new Go$Uint64(0, 1000000), new Go$Uint64(0, 10000000), new Go$Uint64(0, 100000000), new Go$Uint64(0, 1000000000), new Go$Uint64(2, 1410065408), new Go$Uint64(23, 1215752192), new Go$Uint64(232, 3567587328), new Go$Uint64(2328, 1316134912), new Go$Uint64(23283, 276447232), new Go$Uint64(232830, 2764472320), new Go$Uint64(2328306, 1874919424), new Go$Uint64(23283064, 1569325056), new Go$Uint64(232830643, 2808348672), new Go$Uint64(2328306436, 2313682944)]);
		float32info = new floatInfo.Ptr(23, 8, -127);
		float64info = new floatInfo.Ptr(52, 11, -1023);
		isPrint16 = new (go$sliceType(Go$Uint16))([32, 126, 161, 887, 890, 894, 900, 1319, 1329, 1366, 1369, 1418, 1423, 1479, 1488, 1514, 1520, 1524, 1542, 1563, 1566, 1805, 1808, 1866, 1869, 1969, 1984, 2042, 2048, 2093, 2096, 2139, 2142, 2142, 2208, 2220, 2276, 2444, 2447, 2448, 2451, 2482, 2486, 2489, 2492, 2500, 2503, 2504, 2507, 2510, 2519, 2519, 2524, 2531, 2534, 2555, 2561, 2570, 2575, 2576, 2579, 2617, 2620, 2626, 2631, 2632, 2635, 2637, 2641, 2641, 2649, 2654, 2662, 2677, 2689, 2745, 2748, 2765, 2768, 2768, 2784, 2787, 2790, 2801, 2817, 2828, 2831, 2832, 2835, 2873, 2876, 2884, 2887, 2888, 2891, 2893, 2902, 2903, 2908, 2915, 2918, 2935, 2946, 2954, 2958, 2965, 2969, 2975, 2979, 2980, 2984, 2986, 2990, 3001, 3006, 3010, 3014, 3021, 3024, 3024, 3031, 3031, 3046, 3066, 3073, 3129, 3133, 3149, 3157, 3161, 3168, 3171, 3174, 3183, 3192, 3199, 3202, 3257, 3260, 3277, 3285, 3286, 3294, 3299, 3302, 3314, 3330, 3386, 3389, 3406, 3415, 3415, 3424, 3427, 3430, 3445, 3449, 3455, 3458, 3478, 3482, 3517, 3520, 3526, 3530, 3530, 3535, 3551, 3570, 3572, 3585, 3642, 3647, 3675, 3713, 3716, 3719, 3722, 3725, 3725, 3732, 3751, 3754, 3773, 3776, 3789, 3792, 3801, 3804, 3807, 3840, 3948, 3953, 4058, 4096, 4295, 4301, 4301, 4304, 4685, 4688, 4701, 4704, 4749, 4752, 4789, 4792, 4805, 4808, 4885, 4888, 4954, 4957, 4988, 4992, 5017, 5024, 5108, 5120, 5788, 5792, 5872, 5888, 5908, 5920, 5942, 5952, 5971, 5984, 6003, 6016, 6109, 6112, 6121, 6128, 6137, 6144, 6157, 6160, 6169, 6176, 6263, 6272, 6314, 6320, 6389, 6400, 6428, 6432, 6443, 6448, 6459, 6464, 6464, 6468, 6509, 6512, 6516, 6528, 6571, 6576, 6601, 6608, 6618, 6622, 6683, 6686, 6780, 6783, 6793, 6800, 6809, 6816, 6829, 6912, 6987, 6992, 7036, 7040, 7155, 7164, 7223, 7227, 7241, 7245, 7295, 7360, 7367, 7376, 7414, 7424, 7654, 7676, 7957, 7960, 7965, 7968, 8005, 8008, 8013, 8016, 8061, 8064, 8147, 8150, 8175, 8178, 8190, 8208, 8231, 8240, 8286, 8304, 8305, 8308, 8348, 8352, 8378, 8400, 8432, 8448, 8585, 8592, 9203, 9216, 9254, 9280, 9290, 9312, 11084, 11088, 11097, 11264, 11507, 11513, 11559, 11565, 11565, 11568, 11623, 11631, 11632, 11647, 11670, 11680, 11835, 11904, 12019, 12032, 12245, 12272, 12283, 12289, 12438, 12441, 12543, 12549, 12589, 12593, 12730, 12736, 12771, 12784, 19893, 19904, 40908, 40960, 42124, 42128, 42182, 42192, 42539, 42560, 42647, 42655, 42743, 42752, 42899, 42912, 42922, 43000, 43051, 43056, 43065, 43072, 43127, 43136, 43204, 43214, 43225, 43232, 43259, 43264, 43347, 43359, 43388, 43392, 43481, 43486, 43487, 43520, 43574, 43584, 43597, 43600, 43609, 43612, 43643, 43648, 43714, 43739, 43766, 43777, 43782, 43785, 43790, 43793, 43798, 43808, 43822, 43968, 44013, 44016, 44025, 44032, 55203, 55216, 55238, 55243, 55291, 63744, 64109, 64112, 64217, 64256, 64262, 64275, 64279, 64285, 64449, 64467, 64831, 64848, 64911, 64914, 64967, 65008, 65021, 65024, 65049, 65056, 65062, 65072, 65131, 65136, 65276, 65281, 65470, 65474, 65479, 65482, 65487, 65490, 65495, 65498, 65500, 65504, 65518, 65532, 65533]);
		isNotPrint16 = new (go$sliceType(Go$Uint16))([173, 907, 909, 930, 1376, 1416, 1424, 1757, 2111, 2209, 2303, 2424, 2432, 2436, 2473, 2481, 2526, 2564, 2601, 2609, 2612, 2615, 2621, 2653, 2692, 2702, 2706, 2729, 2737, 2740, 2758, 2762, 2820, 2857, 2865, 2868, 2910, 2948, 2961, 2971, 2973, 3017, 3076, 3085, 3089, 3113, 3124, 3141, 3145, 3159, 3204, 3213, 3217, 3241, 3252, 3269, 3273, 3295, 3312, 3332, 3341, 3345, 3397, 3401, 3460, 3506, 3516, 3541, 3543, 3715, 3721, 3736, 3744, 3748, 3750, 3756, 3770, 3781, 3783, 3912, 3992, 4029, 4045, 4294, 4681, 4695, 4697, 4745, 4785, 4799, 4801, 4823, 4881, 5760, 5901, 5997, 6001, 6751, 8024, 8026, 8028, 8030, 8117, 8133, 8156, 8181, 8335, 9984, 11311, 11359, 11558, 11687, 11695, 11703, 11711, 11719, 11727, 11735, 11743, 11930, 12352, 12687, 12831, 13055, 42895, 43470, 43815, 64311, 64317, 64319, 64322, 64325, 65107, 65127, 65141, 65511]);
		isPrint32 = new (go$sliceType(Go$Uint32))([65536, 65613, 65616, 65629, 65664, 65786, 65792, 65794, 65799, 65843, 65847, 65930, 65936, 65947, 66000, 66045, 66176, 66204, 66208, 66256, 66304, 66339, 66352, 66378, 66432, 66499, 66504, 66517, 66560, 66717, 66720, 66729, 67584, 67589, 67592, 67640, 67644, 67644, 67647, 67679, 67840, 67867, 67871, 67897, 67903, 67903, 67968, 68023, 68030, 68031, 68096, 68102, 68108, 68147, 68152, 68154, 68159, 68167, 68176, 68184, 68192, 68223, 68352, 68405, 68409, 68437, 68440, 68466, 68472, 68479, 68608, 68680, 69216, 69246, 69632, 69709, 69714, 69743, 69760, 69825, 69840, 69864, 69872, 69881, 69888, 69955, 70016, 70088, 70096, 70105, 71296, 71351, 71360, 71369, 73728, 74606, 74752, 74850, 74864, 74867, 77824, 78894, 92160, 92728, 93952, 94020, 94032, 94078, 94095, 94111, 110592, 110593, 118784, 119029, 119040, 119078, 119081, 119154, 119163, 119261, 119296, 119365, 119552, 119638, 119648, 119665, 119808, 119967, 119970, 119970, 119973, 119974, 119977, 120074, 120077, 120134, 120138, 120485, 120488, 120779, 120782, 120831, 126464, 126500, 126503, 126523, 126530, 126530, 126535, 126548, 126551, 126564, 126567, 126619, 126625, 126651, 126704, 126705, 126976, 127019, 127024, 127123, 127136, 127150, 127153, 127166, 127169, 127199, 127232, 127242, 127248, 127339, 127344, 127386, 127462, 127490, 127504, 127546, 127552, 127560, 127568, 127569, 127744, 127776, 127792, 127868, 127872, 127891, 127904, 127946, 127968, 127984, 128000, 128252, 128256, 128317, 128320, 128323, 128336, 128359, 128507, 128576, 128581, 128591, 128640, 128709, 128768, 128883, 131072, 173782, 173824, 177972, 177984, 178205, 194560, 195101, 917760, 917999]);
		isNotPrint32 = new (go$sliceType(Go$Uint16))([12, 39, 59, 62, 799, 926, 2057, 2102, 2134, 2564, 2580, 2584, 4285, 4405, 54357, 54429, 54445, 54458, 54460, 54468, 54534, 54549, 54557, 54586, 54591, 54597, 54609, 60932, 60960, 60963, 60968, 60979, 60984, 60986, 61000, 61002, 61004, 61008, 61011, 61016, 61018, 61020, 61022, 61024, 61027, 61035, 61043, 61048, 61053, 61055, 61066, 61092, 61098, 61648, 61743, 62262, 62405, 62527, 62529, 62712]);
		shifts = go$toNativeArray("Uint", [0, 0, 1, 0, 2, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0]);
	}
	return go$pkg;
})();
go$packages["reflect"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], strconv = go$packages["strconv"], sync = go$packages["sync"], math = go$packages["math"], runtime = go$packages["runtime"], mapIter, Type, Kind, rtype, method, uncommonType, ChanDir, arrayType, chanType, funcType, imethod, interfaceType, mapType, ptrType, sliceType, structField, structType, Method, StructField, StructTag, fieldScan, Value, flag, ValueError, iword, jsType, reflectType, isWrapped, copyStruct, zeroVal, makeIndir, makeIndirValue, MakeSlice, jsObject, TypeOf, ValueOf, SliceOf, Zero, unsafe_New, chanclose, chanrecv, chansend, mapaccess, mapassign, mapiterinit, mapiterkey, mapiternext, maplen, makeMethodValue, implements$1, directlyAssignable, haveIdenticalUnderlyingType, toType, methodName, methodReceiver, valueInterface, overflowFloat32, typesMustMatch, New, convertOp, makeInt, makeFloat, makeComplex, makeString, cvtInt, cvtUint, cvtFloatInt, cvtFloatUint, cvtIntFloat, cvtUintFloat, cvtFloat, cvtComplex, cvtIntString, cvtUintString, cvtBytesString, cvtStringBytes, cvtRunesString, cvtStringRunes, cvtDirect, cvtT2I, cvtI2I, call, ifaceE2I, kindNames, uint8Type;
	mapIter = go$pkg.mapIter = go$newType(0, "Struct", "reflect.mapIter", "mapIter", "reflect", function(t_, m_, keys_, i_) {
		this.go$val = this;
		this.t = t_ !== undefined ? t_ : null;
		this.m = m_ !== undefined ? m_ : null;
		this.keys = keys_ !== undefined ? keys_ : null;
		this.i = i_ !== undefined ? i_ : 0;
	});
	Type = go$pkg.Type = go$newType(8, "Interface", "reflect.Type", "Type", "reflect", null);
	Kind = go$pkg.Kind = go$newType(4, "Uint", "reflect.Kind", "Kind", "reflect", null);
	rtype = go$pkg.rtype = go$newType(0, "Struct", "reflect.rtype", "rtype", "reflect", function(size_, hash_, _$2_, align_, fieldAlign_, kind_, alg_, gc_, string_, uncommonType_, ptrToThis_) {
		this.go$val = this;
		this.size = size_ !== undefined ? size_ : 0;
		this.hash = hash_ !== undefined ? hash_ : 0;
		this._$2 = _$2_ !== undefined ? _$2_ : 0;
		this.align = align_ !== undefined ? align_ : 0;
		this.fieldAlign = fieldAlign_ !== undefined ? fieldAlign_ : 0;
		this.kind = kind_ !== undefined ? kind_ : 0;
		this.alg = alg_ !== undefined ? alg_ : (go$ptrType(Go$Uintptr)).nil;
		this.gc = gc_ !== undefined ? gc_ : 0;
		this.string = string_ !== undefined ? string_ : (go$ptrType(Go$String)).nil;
		this.uncommonType = uncommonType_ !== undefined ? uncommonType_ : (go$ptrType(uncommonType)).nil;
		this.ptrToThis = ptrToThis_ !== undefined ? ptrToThis_ : (go$ptrType(rtype)).nil;
	});
	method = go$pkg.method = go$newType(0, "Struct", "reflect.method", "method", "reflect", function(name_, pkgPath_, mtyp_, typ_, ifn_, tfn_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : (go$ptrType(Go$String)).nil;
		this.pkgPath = pkgPath_ !== undefined ? pkgPath_ : (go$ptrType(Go$String)).nil;
		this.mtyp = mtyp_ !== undefined ? mtyp_ : (go$ptrType(rtype)).nil;
		this.typ = typ_ !== undefined ? typ_ : (go$ptrType(rtype)).nil;
		this.ifn = ifn_ !== undefined ? ifn_ : 0;
		this.tfn = tfn_ !== undefined ? tfn_ : 0;
	});
	uncommonType = go$pkg.uncommonType = go$newType(0, "Struct", "reflect.uncommonType", "uncommonType", "reflect", function(name_, pkgPath_, methods_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : (go$ptrType(Go$String)).nil;
		this.pkgPath = pkgPath_ !== undefined ? pkgPath_ : (go$ptrType(Go$String)).nil;
		this.methods = methods_ !== undefined ? methods_ : (go$sliceType(method)).nil;
	});
	ChanDir = go$pkg.ChanDir = go$newType(4, "Int", "reflect.ChanDir", "ChanDir", "reflect", null);
	arrayType = go$pkg.arrayType = go$newType(0, "Struct", "reflect.arrayType", "arrayType", "reflect", function(rtype_, elem_, slice_, len_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.elem = elem_ !== undefined ? elem_ : (go$ptrType(rtype)).nil;
		this.slice = slice_ !== undefined ? slice_ : (go$ptrType(rtype)).nil;
		this.len = len_ !== undefined ? len_ : 0;
	});
	chanType = go$pkg.chanType = go$newType(0, "Struct", "reflect.chanType", "chanType", "reflect", function(rtype_, elem_, dir_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.elem = elem_ !== undefined ? elem_ : (go$ptrType(rtype)).nil;
		this.dir = dir_ !== undefined ? dir_ : 0;
	});
	funcType = go$pkg.funcType = go$newType(0, "Struct", "reflect.funcType", "funcType", "reflect", function(rtype_, dotdotdot_, in$2_, out_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.dotdotdot = dotdotdot_ !== undefined ? dotdotdot_ : false;
		this.in$2 = in$2_ !== undefined ? in$2_ : (go$sliceType((go$ptrType(rtype)))).nil;
		this.out = out_ !== undefined ? out_ : (go$sliceType((go$ptrType(rtype)))).nil;
	});
	imethod = go$pkg.imethod = go$newType(0, "Struct", "reflect.imethod", "imethod", "reflect", function(name_, pkgPath_, typ_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : (go$ptrType(Go$String)).nil;
		this.pkgPath = pkgPath_ !== undefined ? pkgPath_ : (go$ptrType(Go$String)).nil;
		this.typ = typ_ !== undefined ? typ_ : (go$ptrType(rtype)).nil;
	});
	interfaceType = go$pkg.interfaceType = go$newType(0, "Struct", "reflect.interfaceType", "interfaceType", "reflect", function(rtype_, methods_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.methods = methods_ !== undefined ? methods_ : (go$sliceType(imethod)).nil;
	});
	mapType = go$pkg.mapType = go$newType(0, "Struct", "reflect.mapType", "mapType", "reflect", function(rtype_, key_, elem_, bucket_, hmap_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.key = key_ !== undefined ? key_ : (go$ptrType(rtype)).nil;
		this.elem = elem_ !== undefined ? elem_ : (go$ptrType(rtype)).nil;
		this.bucket = bucket_ !== undefined ? bucket_ : (go$ptrType(rtype)).nil;
		this.hmap = hmap_ !== undefined ? hmap_ : (go$ptrType(rtype)).nil;
	});
	ptrType = go$pkg.ptrType = go$newType(0, "Struct", "reflect.ptrType", "ptrType", "reflect", function(rtype_, elem_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.elem = elem_ !== undefined ? elem_ : (go$ptrType(rtype)).nil;
	});
	sliceType = go$pkg.sliceType = go$newType(0, "Struct", "reflect.sliceType", "sliceType", "reflect", function(rtype_, elem_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.elem = elem_ !== undefined ? elem_ : (go$ptrType(rtype)).nil;
	});
	structField = go$pkg.structField = go$newType(0, "Struct", "reflect.structField", "structField", "reflect", function(name_, pkgPath_, typ_, tag_, offset_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : (go$ptrType(Go$String)).nil;
		this.pkgPath = pkgPath_ !== undefined ? pkgPath_ : (go$ptrType(Go$String)).nil;
		this.typ = typ_ !== undefined ? typ_ : (go$ptrType(rtype)).nil;
		this.tag = tag_ !== undefined ? tag_ : (go$ptrType(Go$String)).nil;
		this.offset = offset_ !== undefined ? offset_ : 0;
	});
	structType = go$pkg.structType = go$newType(0, "Struct", "reflect.structType", "structType", "reflect", function(rtype_, fields_) {
		this.go$val = this;
		this.rtype = rtype_ !== undefined ? rtype_ : new rtype.Ptr();
		this.fields = fields_ !== undefined ? fields_ : (go$sliceType(structField)).nil;
	});
	Method = go$pkg.Method = go$newType(0, "Struct", "reflect.Method", "Method", "reflect", function(Name_, PkgPath_, Type_, Func_, Index_) {
		this.go$val = this;
		this.Name = Name_ !== undefined ? Name_ : "";
		this.PkgPath = PkgPath_ !== undefined ? PkgPath_ : "";
		this.Type = Type_ !== undefined ? Type_ : null;
		this.Func = Func_ !== undefined ? Func_ : new Value.Ptr();
		this.Index = Index_ !== undefined ? Index_ : 0;
	});
	StructField = go$pkg.StructField = go$newType(0, "Struct", "reflect.StructField", "StructField", "reflect", function(Name_, PkgPath_, Type_, Tag_, Offset_, Index_, Anonymous_) {
		this.go$val = this;
		this.Name = Name_ !== undefined ? Name_ : "";
		this.PkgPath = PkgPath_ !== undefined ? PkgPath_ : "";
		this.Type = Type_ !== undefined ? Type_ : null;
		this.Tag = Tag_ !== undefined ? Tag_ : "";
		this.Offset = Offset_ !== undefined ? Offset_ : 0;
		this.Index = Index_ !== undefined ? Index_ : (go$sliceType(Go$Int)).nil;
		this.Anonymous = Anonymous_ !== undefined ? Anonymous_ : false;
	});
	StructTag = go$pkg.StructTag = go$newType(8, "String", "reflect.StructTag", "StructTag", "reflect", null);
	fieldScan = go$pkg.fieldScan = go$newType(0, "Struct", "reflect.fieldScan", "fieldScan", "reflect", function(typ_, index_) {
		this.go$val = this;
		this.typ = typ_ !== undefined ? typ_ : (go$ptrType(structType)).nil;
		this.index = index_ !== undefined ? index_ : (go$sliceType(Go$Int)).nil;
	});
	Value = go$pkg.Value = go$newType(0, "Struct", "reflect.Value", "Value", "reflect", function(typ_, val_, flag_) {
		this.go$val = this;
		this.typ = typ_ !== undefined ? typ_ : (go$ptrType(rtype)).nil;
		this.val = val_ !== undefined ? val_ : 0;
		this.flag = flag_ !== undefined ? flag_ : 0;
	});
	flag = go$pkg.flag = go$newType(4, "Uintptr", "reflect.flag", "flag", "reflect", null);
	ValueError = go$pkg.ValueError = go$newType(0, "Struct", "reflect.ValueError", "ValueError", "reflect", function(Method_, Kind_) {
		this.go$val = this;
		this.Method = Method_ !== undefined ? Method_ : "";
		this.Kind = Kind_ !== undefined ? Kind_ : 0;
	});
	iword = go$pkg.iword = go$newType(4, "UnsafePointer", "reflect.iword", "iword", "reflect", null);
	jsType = function(typ) {
		return typ.jsType;
	};
	reflectType = function(typ) {
		var x;
		return (x = go$internalize(typ.reflectType(), go$emptyInterface), (x !== null && x.constructor === (go$ptrType(rtype)) ? x.go$val : go$typeAssertionFailed(x, (go$ptrType(rtype)))));
	};
	isWrapped = function(typ) {
		var _ref;
		_ref = typ.Kind();
		if (_ref === 1 || _ref === 2 || _ref === 3 || _ref === 4 || _ref === 5 || _ref === 7 || _ref === 8 || _ref === 9 || _ref === 10 || _ref === 12 || _ref === 13 || _ref === 14 || _ref === 17 || _ref === 21 || _ref === 19 || _ref === 24 || _ref === 25) {
			return true;
		} else if (_ref === 22) {
			return typ.Elem().Kind() === 17;
		}
		return false;
	};
	copyStruct = function(dst, src, typ) {
		var fields, i, name;
		fields = jsType(typ).fields;
		i = 0;
		while (i < go$parseInt(fields.length)) {
			name = go$internalize(fields[i][0], Go$String);
			dst[go$externalize(name, Go$String)] = src[go$externalize(name, Go$String)];
			i = i + 1 >> 0;
		}
	};
	zeroVal = function(typ) {
		var _ref, elemType;
		_ref = typ.Kind();
		if (_ref === 1) {
			return false;
		} else if (_ref === 2 || _ref === 3 || _ref === 4 || _ref === 5 || _ref === 7 || _ref === 8 || _ref === 9 || _ref === 10 || _ref === 12 || _ref === 13 || _ref === 14) {
			return 0;
		} else if (_ref === 6 || _ref === 11 || _ref === 15 || _ref === 16) {
			return new (jsType(typ))(0, 0);
		} else if (_ref === 17) {
			elemType = typ.Elem();
			return go$makeNativeArray(jsType(elemType).kind, typ.Len(), go$externalize((function() {
				return zeroVal(elemType);
			}), (go$funcType([], [js.Object], false))));
		} else if (_ref === 19) {
			return go$throwNilPointerError;
		} else if (_ref === 20) {
			return null;
		} else if (_ref === 21) {
			return false;
		} else if (_ref === 18 || _ref === 22 || _ref === 23) {
			return jsType(typ).nil;
		} else if (_ref === 24) {
			return "";
		} else if (_ref === 25) {
			return new (jsType(typ).Ptr)();
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Zero", typ.Kind()));
		}
	};
	makeIndir = function(t, v) {
		var rt;
		rt = (t !== null && t.constructor === (go$ptrType(rtype)) ? t.go$val : go$typeAssertionFailed(t, (go$ptrType(rtype))));
		if (rt.size > 4) {
			return go$newDataPointer(v, jsType(rt.ptrTo()));
		}
		return v;
	};
	makeIndirValue = function(t, v) {
		var rt;
		rt = (t !== null && t.constructor === (go$ptrType(rtype)) ? t.go$val : go$typeAssertionFailed(t, (go$ptrType(rtype))));
		return new Value.Ptr(rt, go$newDataPointer(v, jsType(rt.ptrTo())), (2 | ((t.Kind() >>> 0) << 4 >>> 0)) >>> 0);
	};
	MakeSlice = go$pkg.MakeSlice = function(typ, len, cap) {
		var _struct;
		if (!((typ.Kind() === 23))) {
			throw go$panic(new Go$String("reflect.MakeSlice of non-slice type"));
		}
		if (len < 0) {
			throw go$panic(new Go$String("reflect.MakeSlice: negative len"));
		}
		if (cap < 0) {
			throw go$panic(new Go$String("reflect.MakeSlice: negative cap"));
		}
		if (len > cap) {
			throw go$panic(new Go$String("reflect.MakeSlice: len > cap"));
		}
		return (_struct = makeIndirValue(typ, jsType(typ).make(len, cap, go$externalize((function() {
			return zeroVal(typ.Elem());
		}), (go$funcType([], [js.Object], false))))), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	jsObject = function() {
		return reflectType(go$packages[go$externalize("github.com/gopherjs/gopherjs/js", Go$String)].Object);
	};
	TypeOf = go$pkg.TypeOf = function(i) {
		var c;
		if (go$interfaceIsEqual(i, null)) {
			return null;
		}
		c = i.constructor;
		if (c.kind === undefined) {
			return jsObject();
		}
		return reflectType(c);
	};
	ValueOf = go$pkg.ValueOf = function(i) {
		var c, typ, _struct;
		if (go$interfaceIsEqual(i, null)) {
			return new Value.Ptr((go$ptrType(rtype)).nil, 0, 0);
		}
		c = i.constructor;
		if (c.kind === undefined) {
			return new Value.Ptr(jsObject(), i, 320);
		}
		typ = reflectType(c);
		if (!((typ.Kind() === 17)) && !((typ.Kind() === 25)) && typ.Size() > 4) {
			return (_struct = makeIndirValue(typ, i.go$val), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		}
		return new Value.Ptr(typ, i.go$val, (typ.Kind() >>> 0) << 4 >>> 0);
	};
	rtype.Ptr.prototype.ptrTo = function() {
		var t;
		t = this;
		return reflectType(go$ptrType(jsType(t)));
	};
	rtype.prototype.ptrTo = function() { return this.go$val.ptrTo(); };
	SliceOf = go$pkg.SliceOf = function(t) {
		return reflectType(go$sliceType(jsType(t)));
	};
	Zero = go$pkg.Zero = function(typ) {
		return new Value.Ptr((typ !== null && typ.constructor === (go$ptrType(rtype)) ? typ.go$val : go$typeAssertionFailed(typ, (go$ptrType(rtype)))), zeroVal(typ), (typ.Kind() >>> 0) << 4 >>> 0);
	};
	unsafe_New = function(typ) {
		var _ref;
		_ref = typ.Kind();
		if (_ref === 25) {
			return new (jsType(typ).Ptr)();
		} else if (_ref === 17) {
			return zeroVal(typ);
		} else {
			return go$newDataPointer(zeroVal(typ), jsType(typ.ptrTo()));
		}
	};
	chanclose = function(ch) {
		go$notSupported(go$externalize("channels", Go$String));
		throw go$panic(new Go$String("unreachable"));
	};
	chanrecv = function(t, ch, nb) {
		var val, selected, received;
		val = 0;
		selected = false;
		received = false;
		go$notSupported(go$externalize("channels", Go$String));
		throw go$panic(new Go$String("unreachable"));
	};
	chansend = function(t, ch, val, nb) {
		go$notSupported(go$externalize("channels", Go$String));
		throw go$panic(new Go$String("unreachable"));
	};
	mapaccess = function(t, m, key) {
		var val, ok, k, entry, _tmp, _tmp$1, _tmp$2, _tmp$3;
		val = 0;
		ok = false;
		k = key;
		if (!(k.go$key === undefined)) {
			k = k.go$key();
		}
		entry = m[go$externalize(go$internalize(k, Go$String), Go$String)];
		if (entry === undefined) {
			_tmp = 0; _tmp$1 = false; val = _tmp; ok = _tmp$1;
			return [val, ok];
		}
		_tmp$2 = makeIndir(t.Elem(), entry.v); _tmp$3 = true; val = _tmp$2; ok = _tmp$3;
		return [val, ok];
	};
	mapassign = function(t, m, key, val, ok) {
		var k, jsVal, newVal, entry;
		k = key;
		if (!(k.go$key === undefined)) {
			k = k.go$key();
		}
		if (!ok) {
			delete m[go$externalize(go$internalize(k, Go$String), Go$String)];
			return;
		}
		jsVal = val;
		if (t.Elem().Kind() === 25) {
			newVal = new (go$global.Object)();
			copyStruct(newVal, jsVal, t.Elem());
			jsVal = newVal;
		}
		entry = new (go$global.Object)();
		entry.k = go$externalize(key, iword);
		entry.v = jsVal;
		m[go$externalize(go$internalize(k, Go$String), Go$String)] = entry;
	};
	mapiterinit = function(t, m) {
		return new mapIter.Ptr(t, m, go$keys(go$externalize(m, iword)), 0);
	};
	mapiterkey = function(it) {
		var key, ok, iter, k, _tmp, x, _tmp$1;
		key = 0;
		ok = false;
		iter = it;
		k = iter.keys[(go$parseInt(iter.i) >> 0)];
		_tmp = makeIndir((x = go$internalize(iter.t, go$emptyInterface), (x !== null && x.constructor === (go$ptrType(rtype)) ? x.go$val : go$typeAssertionFailed(x, (go$ptrType(rtype))))).Key(), iter.m[go$externalize(go$internalize(k, Go$String), Go$String)].k); _tmp$1 = true; key = _tmp; ok = _tmp$1;
		return [key, ok];
	};
	mapiternext = function(it) {
		var iter;
		iter = it;
		iter.i = (go$parseInt(iter.i) >> 0) + 1 >> 0;
	};
	maplen = function(m) {
		return go$parseInt(go$keys(go$externalize(m, iword)).length);
	};
	Value.Ptr.prototype.iword = function() {
		var _struct, v, val;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (!((((v.flag & 2) >>> 0) === 0)) && !((v.typ.Kind() === 17)) && !((v.typ.Kind() === 25))) {
			val = v.val.go$get();
			if ((v.typ.Kind() === 11) || (v.typ.Kind() === 6)) {
				val = new (jsType(v.typ))(val.high, val.low);
			}
			if ((v.typ.Kind() === 15) || (v.typ.Kind() === 16)) {
				val = new (jsType(v.typ))(val.real, val.imag);
			}
			return val;
		}
		return v.val;
	};
	Value.prototype.iword = function() { return this.go$val.iword(); };
	Value.Ptr.prototype.Cap = function() {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 17) {
			return v.typ.Len();
		} else if (_ref === 23) {
			return go$parseInt(v.iword().capacity) >> 0;
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.Cap", k));
	};
	Value.prototype.Cap = function() { return this.go$val.Cap(); };
	Value.Ptr.prototype.IsNil = function() {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 18 || _ref === 22 || _ref === 23) {
			return v.iword() === jsType(v.typ).nil;
		} else if (_ref === 19) {
			return v.iword() === go$throwNilPointerError;
		} else if (_ref === 21) {
			return v.iword() === false;
		} else if (_ref === 20) {
			return v.iword() === null;
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.IsNil", k));
		}
	};
	Value.prototype.IsNil = function() { return this.go$val.IsNil(); };
	Value.Ptr.prototype.Len = function() {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 17 || _ref === 23 || _ref === 24) {
			return go$parseInt(v.iword().length);
		} else if (_ref === 21) {
			return go$parseInt(go$keys(go$externalize(v.iword(), iword)).length);
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.Len", k));
		}
	};
	Value.prototype.Len = function() { return this.go$val.Len(); };
	Value.Ptr.prototype.Pointer = function() {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 18 || _ref === 21 || _ref === 22 || _ref === 23 || _ref === 26) {
			if (v.IsNil()) {
				return 0;
			}
			return v.iword();
		} else if (_ref === 19) {
			if (v.IsNil()) {
				return 0;
			}
			return 1;
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.Pointer", k));
		}
	};
	Value.prototype.Pointer = function() { return this.go$val.Pointer(); };
	makeMethodValue = function(op, v) {
			if ((v.flag & flagMethod) === 0) {
				throw go$panic(new Go$String("reflect: internal error: invalid use of makePartialFunc"));
			}

			var tuple = methodReceiver(op, v, v.flag >> flagMethodShift);
			var fn = tuple[1];
			var rcvr = tuple[2];
			var fv = function() { return fn.apply(rcvr, arguments); };
			return new Value.Ptr(v.Type(), fv, (v.flag & flagRO) | (Func << flagKindShift));
		};
	Kind.prototype.String = function() {
		var k;
		k = this.go$val;
		if ((k >> 0) < kindNames.length) {
			return ((k < 0 || k >= kindNames.length) ? go$throwRuntimeError("index out of range") : kindNames.array[kindNames.offset + k]);
		}
		return "kind" + strconv.Itoa((k >> 0));
	};
	go$ptrType(Kind).prototype.String = function() { return new Kind(this.go$get()).String(); };
	uncommonType.Ptr.prototype.uncommon = function() {
		var t;
		t = this;
		return t;
	};
	uncommonType.prototype.uncommon = function() { return this.go$val.uncommon(); };
	uncommonType.Ptr.prototype.PkgPath = function() {
		var t;
		t = this;
		if (t === (go$ptrType(uncommonType)).nil || go$pointerIsEqual(t.pkgPath, (go$ptrType(Go$String)).nil)) {
			return "";
		}
		return t.pkgPath.go$get();
	};
	uncommonType.prototype.PkgPath = function() { return this.go$val.PkgPath(); };
	uncommonType.Ptr.prototype.Name = function() {
		var t;
		t = this;
		if (t === (go$ptrType(uncommonType)).nil || go$pointerIsEqual(t.name, (go$ptrType(Go$String)).nil)) {
			return "";
		}
		return t.name.go$get();
	};
	uncommonType.prototype.Name = function() { return this.go$val.Name(); };
	rtype.Ptr.prototype.String = function() {
		var t;
		t = this;
		return t.string.go$get();
	};
	rtype.prototype.String = function() { return this.go$val.String(); };
	rtype.Ptr.prototype.Size = function() {
		var t;
		t = this;
		return t.size;
	};
	rtype.prototype.Size = function() { return this.go$val.Size(); };
	rtype.Ptr.prototype.Bits = function() {
		var t, k, x;
		t = this;
		if (t === (go$ptrType(rtype)).nil) {
			throw go$panic(new Go$String("reflect: Bits of nil Type"));
		}
		k = t.Kind();
		if (k < 2 || k > 16) {
			throw go$panic(new Go$String("reflect: Bits of non-arithmetic Type " + t.String()));
		}
		return (x = (t.size >> 0), (((x >>> 16 << 16) * 8 >> 0) + (x << 16 >>> 16) * 8) >> 0);
	};
	rtype.prototype.Bits = function() { return this.go$val.Bits(); };
	rtype.Ptr.prototype.Align = function() {
		var t;
		t = this;
		return (t.align >> 0);
	};
	rtype.prototype.Align = function() { return this.go$val.Align(); };
	rtype.Ptr.prototype.FieldAlign = function() {
		var t;
		t = this;
		return (t.fieldAlign >> 0);
	};
	rtype.prototype.FieldAlign = function() { return this.go$val.FieldAlign(); };
	rtype.Ptr.prototype.Kind = function() {
		var t;
		t = this;
		return (((t.kind & 127) >>> 0) >>> 0);
	};
	rtype.prototype.Kind = function() { return this.go$val.Kind(); };
	rtype.Ptr.prototype.common = function() {
		var t;
		t = this;
		return t;
	};
	rtype.prototype.common = function() { return this.go$val.common(); };
	uncommonType.Ptr.prototype.Method = function(i) {
			if (this === uncommonType.Ptr.nil || i < 0 || i >= this.methods.length) {
				throw go$panic(new Go$String("reflect: Method index out of range"));
			}
			var p = this.methods.array[i];
			var fl = Func << flagKindShift;
			var pkgPath = "";
			if (p.pkgPath.go$get !== go$throwNilPointerError) {
				pkgPath = p.pkgPath.go$get();
				fl |= flagRO;
			}
			var mt = p.typ;
			var name = p.name.go$get();
			if (go$reservedKeywords.indexOf(name) !== -1) {
				name += "$";
			}
			var fn = function(rcvr) {
				return rcvr[name].apply(rcvr, Go$Array.prototype.slice.apply(arguments, [1]));
			}
			return new Method.Ptr(p.name.go$get(), pkgPath, mt, new Value.Ptr(mt, fn, fl), i);
		};
	uncommonType.prototype.Method = function() { return this.go$val.Method(); };
	uncommonType.Ptr.prototype.NumMethod = function() {
		var t;
		t = this;
		if (t === (go$ptrType(uncommonType)).nil) {
			return 0;
		}
		return t.methods.length;
	};
	uncommonType.prototype.NumMethod = function() { return this.go$val.NumMethod(); };
	uncommonType.Ptr.prototype.MethodByName = function(name) {
		var m, ok, t, _struct, _struct$1, p, _ref, _i, i, x, _tmp, _struct$2, _struct$3, _tmp$1, _struct$4, _struct$5, _struct$6, _struct$7;
		m = new Method.Ptr();
		ok = false;
		t = this;
		if (t === (go$ptrType(uncommonType)).nil) {
			return [(_struct = m, new Method.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, (_struct$1 = _struct.Func, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), _struct.Index)), ok];
		}
		p = (go$ptrType(method)).nil;
		_ref = t.methods;
		_i = 0;
		while (_i < _ref.length) {
			i = _i;
			p = (x = t.methods, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
			if (!(go$pointerIsEqual(p.name, (go$ptrType(Go$String)).nil)) && p.name.go$get() === name) {
				_tmp = (_struct$2 = t.Method(i), new Method.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, (_struct$3 = _struct$2.Func, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), _struct$2.Index)); _tmp$1 = true; m = _tmp; ok = _tmp$1;
				return [(_struct$4 = m, new Method.Ptr(_struct$4.Name, _struct$4.PkgPath, _struct$4.Type, (_struct$5 = _struct$4.Func, new Value.Ptr(_struct$5.typ, _struct$5.val, _struct$5.flag)), _struct$4.Index)), ok];
			}
			_i++;
		}
		return [(_struct$6 = m, new Method.Ptr(_struct$6.Name, _struct$6.PkgPath, _struct$6.Type, (_struct$7 = _struct$6.Func, new Value.Ptr(_struct$7.typ, _struct$7.val, _struct$7.flag)), _struct$6.Index)), ok];
	};
	uncommonType.prototype.MethodByName = function(name) { return this.go$val.MethodByName(name); };
	rtype.Ptr.prototype.NumMethod = function() {
		var t, tt;
		t = this;
		if (t.Kind() === 20) {
			tt = t.interfaceType;
			return tt.NumMethod();
		}
		return t.uncommonType.NumMethod();
	};
	rtype.prototype.NumMethod = function() { return this.go$val.NumMethod(); };
	rtype.Ptr.prototype.Method = function(i) {
		var m, t, tt, _struct, _struct$1, _struct$2, _struct$3, _struct$4, _struct$5, _struct$6, _struct$7;
		m = new Method.Ptr();
		t = this;
		if (t.Kind() === 20) {
			tt = t.interfaceType;
			m = (_struct = tt.Method(i), new Method.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, (_struct$1 = _struct.Func, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), _struct.Index));
			return (_struct$2 = m, new Method.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, (_struct$3 = _struct$2.Func, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), _struct$2.Index));
		}
		m = (_struct$4 = t.uncommonType.Method(i), new Method.Ptr(_struct$4.Name, _struct$4.PkgPath, _struct$4.Type, (_struct$5 = _struct$4.Func, new Value.Ptr(_struct$5.typ, _struct$5.val, _struct$5.flag)), _struct$4.Index));
		return (_struct$6 = m, new Method.Ptr(_struct$6.Name, _struct$6.PkgPath, _struct$6.Type, (_struct$7 = _struct$6.Func, new Value.Ptr(_struct$7.typ, _struct$7.val, _struct$7.flag)), _struct$6.Index));
	};
	rtype.prototype.Method = function(i) { return this.go$val.Method(i); };
	rtype.Ptr.prototype.MethodByName = function(name) {
		var m, ok, t, tt, _tuple, _struct, _struct$1, _struct$2, _struct$3, _tuple$1, _struct$4, _struct$5, _struct$6, _struct$7;
		m = new Method.Ptr();
		ok = false;
		t = this;
		if (t.Kind() === 20) {
			tt = t.interfaceType;
			_tuple = tt.MethodByName(name); m = (_struct = _tuple[0], new Method.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, (_struct$1 = _struct.Func, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), _struct.Index)); ok = _tuple[1];
			return [(_struct$2 = m, new Method.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, (_struct$3 = _struct$2.Func, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), _struct$2.Index)), ok];
		}
		_tuple$1 = t.uncommonType.MethodByName(name); m = (_struct$4 = _tuple$1[0], new Method.Ptr(_struct$4.Name, _struct$4.PkgPath, _struct$4.Type, (_struct$5 = _struct$4.Func, new Value.Ptr(_struct$5.typ, _struct$5.val, _struct$5.flag)), _struct$4.Index)); ok = _tuple$1[1];
		return [(_struct$6 = m, new Method.Ptr(_struct$6.Name, _struct$6.PkgPath, _struct$6.Type, (_struct$7 = _struct$6.Func, new Value.Ptr(_struct$7.typ, _struct$7.val, _struct$7.flag)), _struct$6.Index)), ok];
	};
	rtype.prototype.MethodByName = function(name) { return this.go$val.MethodByName(name); };
	rtype.Ptr.prototype.PkgPath = function() {
		var t;
		t = this;
		return t.uncommonType.PkgPath();
	};
	rtype.prototype.PkgPath = function() { return this.go$val.PkgPath(); };
	rtype.Ptr.prototype.Name = function() {
		var t;
		t = this;
		return t.uncommonType.Name();
	};
	rtype.prototype.Name = function() { return this.go$val.Name(); };
	rtype.Ptr.prototype.ChanDir = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 18))) {
			throw go$panic(new Go$String("reflect: ChanDir of non-chan type"));
		}
		tt = t.chanType;
		return (tt.dir >> 0);
	};
	rtype.prototype.ChanDir = function() { return this.go$val.ChanDir(); };
	rtype.Ptr.prototype.IsVariadic = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 19))) {
			throw go$panic(new Go$String("reflect: IsVariadic of non-func type"));
		}
		tt = t.funcType;
		return tt.dotdotdot;
	};
	rtype.prototype.IsVariadic = function() { return this.go$val.IsVariadic(); };
	rtype.Ptr.prototype.Elem = function() {
		var t, _ref, tt, tt$1, tt$2, tt$3, tt$4;
		t = this;
		_ref = t.Kind();
		if (_ref === 17) {
			tt = t.arrayType;
			return toType(tt.elem);
		} else if (_ref === 18) {
			tt$1 = t.chanType;
			return toType(tt$1.elem);
		} else if (_ref === 21) {
			tt$2 = t.mapType;
			return toType(tt$2.elem);
		} else if (_ref === 22) {
			tt$3 = t.ptrType;
			return toType(tt$3.elem);
		} else if (_ref === 23) {
			tt$4 = t.sliceType;
			return toType(tt$4.elem);
		}
		throw go$panic(new Go$String("reflect: Elem of invalid type"));
	};
	rtype.prototype.Elem = function() { return this.go$val.Elem(); };
	rtype.Ptr.prototype.Field = function(i) {
		var t, tt, _struct;
		t = this;
		if (!((t.Kind() === 25))) {
			throw go$panic(new Go$String("reflect: Field of non-struct type"));
		}
		tt = t.structType;
		return (_struct = tt.Field(i), new StructField.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, _struct.Tag, _struct.Offset, _struct.Index, _struct.Anonymous));
	};
	rtype.prototype.Field = function(i) { return this.go$val.Field(i); };
	rtype.Ptr.prototype.FieldByIndex = function(index) {
		var t, tt, _struct;
		t = this;
		if (!((t.Kind() === 25))) {
			throw go$panic(new Go$String("reflect: FieldByIndex of non-struct type"));
		}
		tt = t.structType;
		return (_struct = tt.FieldByIndex(index), new StructField.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, _struct.Tag, _struct.Offset, _struct.Index, _struct.Anonymous));
	};
	rtype.prototype.FieldByIndex = function(index) { return this.go$val.FieldByIndex(index); };
	rtype.Ptr.prototype.FieldByName = function(name) {
		var t, tt;
		t = this;
		if (!((t.Kind() === 25))) {
			throw go$panic(new Go$String("reflect: FieldByName of non-struct type"));
		}
		tt = t.structType;
		return tt.FieldByName(name);
	};
	rtype.prototype.FieldByName = function(name) { return this.go$val.FieldByName(name); };
	rtype.Ptr.prototype.FieldByNameFunc = function(match) {
		var t, tt;
		t = this;
		if (!((t.Kind() === 25))) {
			throw go$panic(new Go$String("reflect: FieldByNameFunc of non-struct type"));
		}
		tt = t.structType;
		return tt.FieldByNameFunc(match);
	};
	rtype.prototype.FieldByNameFunc = function(match) { return this.go$val.FieldByNameFunc(match); };
	rtype.Ptr.prototype.In = function(i) {
		var t, tt, x;
		t = this;
		if (!((t.Kind() === 19))) {
			throw go$panic(new Go$String("reflect: In of non-func type"));
		}
		tt = t.funcType;
		return toType((x = tt.in$2, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i])));
	};
	rtype.prototype.In = function(i) { return this.go$val.In(i); };
	rtype.Ptr.prototype.Key = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 21))) {
			throw go$panic(new Go$String("reflect: Key of non-map type"));
		}
		tt = t.mapType;
		return toType(tt.key);
	};
	rtype.prototype.Key = function() { return this.go$val.Key(); };
	rtype.Ptr.prototype.Len = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 17))) {
			throw go$panic(new Go$String("reflect: Len of non-array type"));
		}
		tt = t.arrayType;
		return (tt.len >> 0);
	};
	rtype.prototype.Len = function() { return this.go$val.Len(); };
	rtype.Ptr.prototype.NumField = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 25))) {
			throw go$panic(new Go$String("reflect: NumField of non-struct type"));
		}
		tt = t.structType;
		return tt.fields.length;
	};
	rtype.prototype.NumField = function() { return this.go$val.NumField(); };
	rtype.Ptr.prototype.NumIn = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 19))) {
			throw go$panic(new Go$String("reflect: NumIn of non-func type"));
		}
		tt = t.funcType;
		return tt.in$2.length;
	};
	rtype.prototype.NumIn = function() { return this.go$val.NumIn(); };
	rtype.Ptr.prototype.NumOut = function() {
		var t, tt;
		t = this;
		if (!((t.Kind() === 19))) {
			throw go$panic(new Go$String("reflect: NumOut of non-func type"));
		}
		tt = t.funcType;
		return tt.out.length;
	};
	rtype.prototype.NumOut = function() { return this.go$val.NumOut(); };
	rtype.Ptr.prototype.Out = function(i) {
		var t, tt, x;
		t = this;
		if (!((t.Kind() === 19))) {
			throw go$panic(new Go$String("reflect: Out of non-func type"));
		}
		tt = t.funcType;
		return toType((x = tt.out, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i])));
	};
	rtype.prototype.Out = function(i) { return this.go$val.Out(i); };
	ChanDir.prototype.String = function() {
		var d, _ref;
		d = this.go$val;
		_ref = d;
		if (_ref === 2) {
			return "chan<-";
		} else if (_ref === 1) {
			return "<-chan";
		} else if (_ref === 3) {
			return "chan";
		}
		return "ChanDir" + strconv.Itoa((d >> 0));
	};
	go$ptrType(ChanDir).prototype.String = function() { return new ChanDir(this.go$get()).String(); };
	interfaceType.Ptr.prototype.Method = function(i) {
		var m, t, _struct, _struct$1, x, p, _struct$2, _struct$3;
		m = new Method.Ptr();
		t = this;
		if (i < 0 || i >= t.methods.length) {
			return (_struct = m, new Method.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, (_struct$1 = _struct.Func, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), _struct.Index));
		}
		p = (x = t.methods, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
		m.Name = p.name.go$get();
		if (!(go$pointerIsEqual(p.pkgPath, (go$ptrType(Go$String)).nil))) {
			m.PkgPath = p.pkgPath.go$get();
		}
		m.Type = toType(p.typ);
		m.Index = i;
		return (_struct$2 = m, new Method.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, (_struct$3 = _struct$2.Func, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), _struct$2.Index));
	};
	interfaceType.prototype.Method = function(i) { return this.go$val.Method(i); };
	interfaceType.Ptr.prototype.NumMethod = function() {
		var t;
		t = this;
		return t.methods.length;
	};
	interfaceType.prototype.NumMethod = function() { return this.go$val.NumMethod(); };
	interfaceType.Ptr.prototype.MethodByName = function(name) {
		var m, ok, t, _struct, _struct$1, p, _ref, _i, i, x, _tmp, _struct$2, _struct$3, _tmp$1, _struct$4, _struct$5, _struct$6, _struct$7;
		m = new Method.Ptr();
		ok = false;
		t = this;
		if (t === (go$ptrType(interfaceType)).nil) {
			return [(_struct = m, new Method.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, (_struct$1 = _struct.Func, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), _struct.Index)), ok];
		}
		p = (go$ptrType(imethod)).nil;
		_ref = t.methods;
		_i = 0;
		while (_i < _ref.length) {
			i = _i;
			p = (x = t.methods, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
			if (p.name.go$get() === name) {
				_tmp = (_struct$2 = t.Method(i), new Method.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, (_struct$3 = _struct$2.Func, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), _struct$2.Index)); _tmp$1 = true; m = _tmp; ok = _tmp$1;
				return [(_struct$4 = m, new Method.Ptr(_struct$4.Name, _struct$4.PkgPath, _struct$4.Type, (_struct$5 = _struct$4.Func, new Value.Ptr(_struct$5.typ, _struct$5.val, _struct$5.flag)), _struct$4.Index)), ok];
			}
			_i++;
		}
		return [(_struct$6 = m, new Method.Ptr(_struct$6.Name, _struct$6.PkgPath, _struct$6.Type, (_struct$7 = _struct$6.Func, new Value.Ptr(_struct$7.typ, _struct$7.val, _struct$7.flag)), _struct$6.Index)), ok];
	};
	interfaceType.prototype.MethodByName = function(name) { return this.go$val.MethodByName(name); };
	StructTag.prototype.Get = function(key) {
		var tag, i, name, qvalue, _tuple, value;
		tag = this.go$val;
		while (!(tag === "")) {
			i = 0;
			while (i < tag.length && (tag.charCodeAt(i) === 32)) {
				i = i + 1 >> 0;
			}
			tag = tag.substring(i);
			if (tag === "") {
				break;
			}
			i = 0;
			while (i < tag.length && !((tag.charCodeAt(i) === 32)) && !((tag.charCodeAt(i) === 58)) && !((tag.charCodeAt(i) === 34))) {
				i = i + 1 >> 0;
			}
			if ((i + 1 >> 0) >= tag.length || !((tag.charCodeAt(i) === 58)) || !((tag.charCodeAt((i + 1 >> 0)) === 34))) {
				break;
			}
			name = tag.substring(0, i);
			tag = tag.substring((i + 1 >> 0));
			i = 1;
			while (i < tag.length && !((tag.charCodeAt(i) === 34))) {
				if (tag.charCodeAt(i) === 92) {
					i = i + 1 >> 0;
				}
				i = i + 1 >> 0;
			}
			if (i >= tag.length) {
				break;
			}
			qvalue = tag.substring(0, (i + 1 >> 0));
			tag = tag.substring((i + 1 >> 0));
			if (key === name) {
				_tuple = strconv.Unquote(qvalue); value = _tuple[0];
				return value;
			}
		}
		return "";
	};
	go$ptrType(StructTag).prototype.Get = function(key) { return new StructTag(this.go$get()).Get(key); };
	structType.Ptr.prototype.Field = function(i) {
		var f, t, _struct, x, p, t$1, _struct$1;
		f = new StructField.Ptr();
		t = this;
		if (i < 0 || i >= t.fields.length) {
			return (_struct = f, new StructField.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, _struct.Tag, _struct.Offset, _struct.Index, _struct.Anonymous));
		}
		p = (x = t.fields, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
		f.Type = toType(p.typ);
		if (!(go$pointerIsEqual(p.name, (go$ptrType(Go$String)).nil))) {
			f.Name = p.name.go$get();
		} else {
			t$1 = f.Type;
			if (t$1.Kind() === 22) {
				t$1 = t$1.Elem();
			}
			f.Name = t$1.Name();
			f.Anonymous = true;
		}
		if (!(go$pointerIsEqual(p.pkgPath, (go$ptrType(Go$String)).nil))) {
			f.PkgPath = p.pkgPath.go$get();
		}
		if (!(go$pointerIsEqual(p.tag, (go$ptrType(Go$String)).nil))) {
			f.Tag = p.tag.go$get();
		}
		f.Offset = p.offset;
		f.Index = new (go$sliceType(Go$Int))([i]);
		return (_struct$1 = f, new StructField.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, _struct$1.Tag, _struct$1.Offset, _struct$1.Index, _struct$1.Anonymous));
	};
	structType.prototype.Field = function(i) { return this.go$val.Field(i); };
	structType.Ptr.prototype.FieldByIndex = function(index) {
		var f, t, _ref, _i, x, i, ft, _struct, _struct$1;
		f = new StructField.Ptr();
		t = this;
		f.Type = toType(t.rtype);
		_ref = index;
		_i = 0;
		while (_i < _ref.length) {
			x = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if (i > 0) {
				ft = f.Type;
				if ((ft.Kind() === 22) && (ft.Elem().Kind() === 25)) {
					ft = ft.Elem();
				}
				f.Type = ft;
			}
			f = (_struct = f.Type.Field(x), new StructField.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, _struct.Tag, _struct.Offset, _struct.Index, _struct.Anonymous));
			_i++;
		}
		return (_struct$1 = f, new StructField.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, _struct$1.Tag, _struct$1.Offset, _struct$1.Index, _struct$1.Anonymous));
	};
	structType.prototype.FieldByIndex = function(index) { return this.go$val.FieldByIndex(index); };
	structType.Ptr.prototype.FieldByNameFunc = function(match) {
		var result, ok, t, current, next, nextCount, _map, _key, visited, _tmp, _tmp$1, count, _ref, _i, _struct, scan, t$1, _entry, _key$1, _ref$1, _i$1, i, x, f, fname, ntyp, _entry$1, _tmp$2, _tmp$3, _struct$1, _struct$2, styp, _entry$2, _key$2, _map$1, _key$3, _key$4, _entry$3, _key$5, index, _struct$3;
		result = new StructField.Ptr();
		ok = false;
		t = this;
		current = new (go$sliceType(fieldScan))([]);
		next = new (go$sliceType(fieldScan))([new fieldScan.Ptr(t, (go$sliceType(Go$Int)).nil)]);
		nextCount = false;
		visited = (_map = new Go$Map(), _map);
		while (next.length > 0) {
			_tmp = next; _tmp$1 = go$subslice(current, 0, 0); current = _tmp; next = _tmp$1;
			count = nextCount;
			nextCount = false;
			_ref = current;
			_i = 0;
			while (_i < _ref.length) {
				scan = (_struct = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]), new fieldScan.Ptr(_struct.typ, _struct.index));
				t$1 = scan.typ;
				if ((_entry = visited[t$1.go$key()], _entry !== undefined ? _entry.v : false)) {
					_i++;
					continue;
				}
				_key$1 = t$1; (visited || go$throwRuntimeError("assignment to entry in nil map"))[_key$1.go$key()] = { k: _key$1, v: true };
				_ref$1 = t$1.fields;
				_i$1 = 0;
				while (_i$1 < _ref$1.length) {
					i = _i$1;
					f = (x = t$1.fields, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
					fname = "";
					ntyp = (go$ptrType(rtype)).nil;
					if (!(go$pointerIsEqual(f.name, (go$ptrType(Go$String)).nil))) {
						fname = f.name.go$get();
					} else {
						ntyp = f.typ;
						if (ntyp.Kind() === 22) {
							ntyp = ntyp.Elem().common();
						}
						fname = ntyp.Name();
					}
					if (match(fname)) {
						if ((_entry$1 = count[t$1.go$key()], _entry$1 !== undefined ? _entry$1.v : 0) > 1 || ok) {
							_tmp$2 = new StructField.Ptr("", "", null, "", 0, (go$sliceType(Go$Int)).nil, false); _tmp$3 = false; result = _tmp$2; ok = _tmp$3;
							return [(_struct$1 = result, new StructField.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, _struct$1.Tag, _struct$1.Offset, _struct$1.Index, _struct$1.Anonymous)), ok];
						}
						result = (_struct$2 = t$1.Field(i), new StructField.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, _struct$2.Tag, _struct$2.Offset, _struct$2.Index, _struct$2.Anonymous));
						result.Index = (go$sliceType(Go$Int)).nil;
						result.Index = go$appendSlice(result.Index, scan.index);
						result.Index = go$append(result.Index, i);
						ok = true;
						_i$1++;
						continue;
					}
					if (ok || ntyp === (go$ptrType(rtype)).nil || !((ntyp.Kind() === 25))) {
						_i$1++;
						continue;
					}
					styp = ntyp.structType;
					if ((_entry$2 = nextCount[styp.go$key()], _entry$2 !== undefined ? _entry$2.v : 0) > 0) {
						_key$2 = styp; (nextCount || go$throwRuntimeError("assignment to entry in nil map"))[_key$2.go$key()] = { k: _key$2, v: 2 };
						_i$1++;
						continue;
					}
					if (nextCount === false) {
						nextCount = (_map$1 = new Go$Map(), _map$1);
					}
					_key$4 = styp; (nextCount || go$throwRuntimeError("assignment to entry in nil map"))[_key$4.go$key()] = { k: _key$4, v: 1 };
					if ((_entry$3 = count[t$1.go$key()], _entry$3 !== undefined ? _entry$3.v : 0) > 1) {
						_key$5 = styp; (nextCount || go$throwRuntimeError("assignment to entry in nil map"))[_key$5.go$key()] = { k: _key$5, v: 2 };
					}
					index = (go$sliceType(Go$Int)).nil;
					index = go$appendSlice(index, scan.index);
					index = go$append(index, i);
					next = go$append(next, new fieldScan.Ptr(styp, index));
					_i$1++;
				}
				_i++;
			}
			if (ok) {
				break;
			}
		}
		return [(_struct$3 = result, new StructField.Ptr(_struct$3.Name, _struct$3.PkgPath, _struct$3.Type, _struct$3.Tag, _struct$3.Offset, _struct$3.Index, _struct$3.Anonymous)), ok];
	};
	structType.prototype.FieldByNameFunc = function(match) { return this.go$val.FieldByNameFunc(match); };
	structType.Ptr.prototype.FieldByName = function(name) {
		var f, present, t, hasAnon, _ref, _i, i, x, tf, _tmp, _struct, _tmp$1, _struct$1, _struct$2, _tuple, _struct$3, _struct$4;
		f = new StructField.Ptr();
		present = false;
		t = this;
		hasAnon = false;
		if (!(name === "")) {
			_ref = t.fields;
			_i = 0;
			while (_i < _ref.length) {
				i = _i;
				tf = (x = t.fields, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
				if (go$pointerIsEqual(tf.name, (go$ptrType(Go$String)).nil)) {
					hasAnon = true;
					_i++;
					continue;
				}
				if (tf.name.go$get() === name) {
					_tmp = (_struct = t.Field(i), new StructField.Ptr(_struct.Name, _struct.PkgPath, _struct.Type, _struct.Tag, _struct.Offset, _struct.Index, _struct.Anonymous)); _tmp$1 = true; f = _tmp; present = _tmp$1;
					return [(_struct$1 = f, new StructField.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, _struct$1.Tag, _struct$1.Offset, _struct$1.Index, _struct$1.Anonymous)), present];
				}
				_i++;
			}
		}
		if (!hasAnon) {
			return [(_struct$2 = f, new StructField.Ptr(_struct$2.Name, _struct$2.PkgPath, _struct$2.Type, _struct$2.Tag, _struct$2.Offset, _struct$2.Index, _struct$2.Anonymous)), present];
		}
		_tuple = t.FieldByNameFunc((function(s) {
			return s === name;
		})); f = (_struct$3 = _tuple[0], new StructField.Ptr(_struct$3.Name, _struct$3.PkgPath, _struct$3.Type, _struct$3.Tag, _struct$3.Offset, _struct$3.Index, _struct$3.Anonymous)); present = _tuple[1];
		return [(_struct$4 = f, new StructField.Ptr(_struct$4.Name, _struct$4.PkgPath, _struct$4.Type, _struct$4.Tag, _struct$4.Offset, _struct$4.Index, _struct$4.Anonymous)), present];
	};
	structType.prototype.FieldByName = function(name) { return this.go$val.FieldByName(name); };
	rtype.Ptr.prototype.Implements = function(u) {
		var t;
		t = this;
		if (go$interfaceIsEqual(u, null)) {
			throw go$panic(new Go$String("reflect: nil type passed to Type.Implements"));
		}
		if (!((u.Kind() === 20))) {
			throw go$panic(new Go$String("reflect: non-interface type passed to Type.Implements"));
		}
		return implements$1((u !== null && u.constructor === (go$ptrType(rtype)) ? u.go$val : go$typeAssertionFailed(u, (go$ptrType(rtype)))), t);
	};
	rtype.prototype.Implements = function(u) { return this.go$val.Implements(u); };
	rtype.Ptr.prototype.AssignableTo = function(u) {
		var t, uu;
		t = this;
		if (go$interfaceIsEqual(u, null)) {
			throw go$panic(new Go$String("reflect: nil type passed to Type.AssignableTo"));
		}
		uu = (u !== null && u.constructor === (go$ptrType(rtype)) ? u.go$val : go$typeAssertionFailed(u, (go$ptrType(rtype))));
		return directlyAssignable(uu, t) || implements$1(uu, t);
	};
	rtype.prototype.AssignableTo = function(u) { return this.go$val.AssignableTo(u); };
	rtype.Ptr.prototype.ConvertibleTo = function(u) {
		var t, uu;
		t = this;
		if (go$interfaceIsEqual(u, null)) {
			throw go$panic(new Go$String("reflect: nil type passed to Type.ConvertibleTo"));
		}
		uu = (u !== null && u.constructor === (go$ptrType(rtype)) ? u.go$val : go$typeAssertionFailed(u, (go$ptrType(rtype))));
		return !(convertOp(uu, t) === go$throwNilPointerError);
	};
	rtype.prototype.ConvertibleTo = function(u) { return this.go$val.ConvertibleTo(u); };
	implements$1 = function(T, V) {
		var t, v, i, j, x, tm, x$1, vm, v$1, i$1, j$1, x$2, tm$1, x$3, vm$1;
		if (!((T.Kind() === 20))) {
			return false;
		}
		t = T.interfaceType;
		if (t.methods.length === 0) {
			return true;
		}
		if (V.Kind() === 20) {
			v = V.interfaceType;
			i = 0;
			j = 0;
			while (j < v.methods.length) {
				tm = (x = t.methods, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
				vm = (x$1 = v.methods, ((j < 0 || j >= x$1.length) ? go$throwRuntimeError("index out of range") : x$1.array[x$1.offset + j]));
				if (go$pointerIsEqual(vm.name, tm.name) && go$pointerIsEqual(vm.pkgPath, tm.pkgPath) && vm.typ === tm.typ) {
					i = i + 1 >> 0;
					if (i >= t.methods.length) {
						return true;
					}
				}
				j = j + 1 >> 0;
			}
			return false;
		}
		v$1 = V.uncommonType.uncommon();
		if (v$1 === (go$ptrType(uncommonType)).nil) {
			return false;
		}
		i$1 = 0;
		j$1 = 0;
		while (j$1 < v$1.methods.length) {
			tm$1 = (x$2 = t.methods, ((i$1 < 0 || i$1 >= x$2.length) ? go$throwRuntimeError("index out of range") : x$2.array[x$2.offset + i$1]));
			vm$1 = (x$3 = v$1.methods, ((j$1 < 0 || j$1 >= x$3.length) ? go$throwRuntimeError("index out of range") : x$3.array[x$3.offset + j$1]));
			if (go$pointerIsEqual(vm$1.name, tm$1.name) && go$pointerIsEqual(vm$1.pkgPath, tm$1.pkgPath) && vm$1.mtyp === tm$1.typ) {
				i$1 = i$1 + 1 >> 0;
				if (i$1 >= t.methods.length) {
					return true;
				}
			}
			j$1 = j$1 + 1 >> 0;
		}
		return false;
	};
	directlyAssignable = function(T, V) {
		if (T === V) {
			return true;
		}
		if (!(T.Name() === "") && !(V.Name() === "") || !((T.Kind() === V.Kind()))) {
			return false;
		}
		return haveIdenticalUnderlyingType(T, V);
	};
	haveIdenticalUnderlyingType = function(T, V) {
		var kind, _ref, t, v, _ref$1, _i, typ, i, x, _ref$2, _i$1, typ$1, i$1, x$1, t$1, v$1, t$2, v$2, _ref$3, _i$2, i$2, x$2, tf, x$3, vf;
		if (T === V) {
			return true;
		}
		kind = T.Kind();
		if (!((kind === V.Kind()))) {
			return false;
		}
		if (1 <= kind && kind <= 16 || (kind === 24) || (kind === 26)) {
			return true;
		}
		_ref = kind;
		if (_ref === 17) {
			return go$interfaceIsEqual(T.Elem(), V.Elem()) && (T.Len() === V.Len());
		} else if (_ref === 18) {
			if ((V.ChanDir() === 3) && go$interfaceIsEqual(T.Elem(), V.Elem())) {
				return true;
			}
			return (V.ChanDir() === T.ChanDir()) && go$interfaceIsEqual(T.Elem(), V.Elem());
		} else if (_ref === 19) {
			t = T.funcType;
			v = V.funcType;
			if (!(t.dotdotdot === v.dotdotdot) || !((t.in$2.length === v.in$2.length)) || !((t.out.length === v.out.length))) {
				return false;
			}
			_ref$1 = t.in$2;
			_i = 0;
			while (_i < _ref$1.length) {
				typ = ((_i < 0 || _i >= _ref$1.length) ? go$throwRuntimeError("index out of range") : _ref$1.array[_ref$1.offset + _i]);
				i = _i;
				if (!(typ === (x = v.in$2, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i])))) {
					return false;
				}
				_i++;
			}
			_ref$2 = t.out;
			_i$1 = 0;
			while (_i$1 < _ref$2.length) {
				typ$1 = ((_i$1 < 0 || _i$1 >= _ref$2.length) ? go$throwRuntimeError("index out of range") : _ref$2.array[_ref$2.offset + _i$1]);
				i$1 = _i$1;
				if (!(typ$1 === (x$1 = v.out, ((i$1 < 0 || i$1 >= x$1.length) ? go$throwRuntimeError("index out of range") : x$1.array[x$1.offset + i$1])))) {
					return false;
				}
				_i$1++;
			}
			return true;
		} else if (_ref === 20) {
			t$1 = T.interfaceType;
			v$1 = V.interfaceType;
			if ((t$1.methods.length === 0) && (v$1.methods.length === 0)) {
				return true;
			}
			return false;
		} else if (_ref === 21) {
			return go$interfaceIsEqual(T.Key(), V.Key()) && go$interfaceIsEqual(T.Elem(), V.Elem());
		} else if (_ref === 22 || _ref === 23) {
			return go$interfaceIsEqual(T.Elem(), V.Elem());
		} else if (_ref === 25) {
			t$2 = T.structType;
			v$2 = V.structType;
			if (!((t$2.fields.length === v$2.fields.length))) {
				return false;
			}
			_ref$3 = t$2.fields;
			_i$2 = 0;
			while (_i$2 < _ref$3.length) {
				i$2 = _i$2;
				tf = (x$2 = t$2.fields, ((i$2 < 0 || i$2 >= x$2.length) ? go$throwRuntimeError("index out of range") : x$2.array[x$2.offset + i$2]));
				vf = (x$3 = v$2.fields, ((i$2 < 0 || i$2 >= x$3.length) ? go$throwRuntimeError("index out of range") : x$3.array[x$3.offset + i$2]));
				if (!(go$pointerIsEqual(tf.name, vf.name)) && (go$pointerIsEqual(tf.name, (go$ptrType(Go$String)).nil) || go$pointerIsEqual(vf.name, (go$ptrType(Go$String)).nil) || !(tf.name.go$get() === vf.name.go$get()))) {
					return false;
				}
				if (!(go$pointerIsEqual(tf.pkgPath, vf.pkgPath)) && (go$pointerIsEqual(tf.pkgPath, (go$ptrType(Go$String)).nil) || go$pointerIsEqual(vf.pkgPath, (go$ptrType(Go$String)).nil) || !(tf.pkgPath.go$get() === vf.pkgPath.go$get()))) {
					return false;
				}
				if (!(tf.typ === vf.typ)) {
					return false;
				}
				if (!(go$pointerIsEqual(tf.tag, vf.tag)) && (go$pointerIsEqual(tf.tag, (go$ptrType(Go$String)).nil) || go$pointerIsEqual(vf.tag, (go$ptrType(Go$String)).nil) || !(tf.tag.go$get() === vf.tag.go$get()))) {
					return false;
				}
				if (!((tf.offset === vf.offset))) {
					return false;
				}
				_i$2++;
			}
			return true;
		}
		return false;
	};
	toType = function(t) {
		if (t === (go$ptrType(rtype)).nil) {
			return null;
		}
		return t;
	};
	flag.prototype.kind = function() {
		var f;
		f = this.go$val;
		return (((((f >>> 4 >>> 0)) & 31) >>> 0) >>> 0);
	};
	go$ptrType(flag).prototype.kind = function() { return new flag(this.go$get()).kind(); };
	ValueError.Ptr.prototype.Error = function() {
		var e;
		e = this;
		if (e.Kind === 0) {
			return "reflect: call of " + e.Method + " on zero Value";
		}
		return "reflect: call of " + e.Method + " on " + (new Kind(e.Kind)).String() + " Value";
	};
	ValueError.prototype.Error = function() { return this.go$val.Error(); };
	methodName = function() {
			return "?FIXME?";
		};
	flag.prototype.mustBe = function(expected) {
		var f, k;
		f = this.go$val;
		k = (new flag(f)).kind();
		if (!((k === expected))) {
			throw go$panic(new ValueError.Ptr(methodName(), k));
		}
	};
	go$ptrType(flag).prototype.mustBe = function(expected) { return new flag(this.go$get()).mustBe(expected); };
	flag.prototype.mustBeExported = function() {
		var f;
		f = this.go$val;
		if (f === 0) {
			throw go$panic(new ValueError.Ptr(methodName(), 0));
		}
		if (!((((f & 1) >>> 0) === 0))) {
			throw go$panic(new Go$String("reflect: " + methodName() + " using value obtained using unexported field"));
		}
	};
	go$ptrType(flag).prototype.mustBeExported = function() { return new flag(this.go$get()).mustBeExported(); };
	flag.prototype.mustBeAssignable = function() {
		var f;
		f = this.go$val;
		if (f === 0) {
			throw go$panic(new ValueError.Ptr(methodName(), 0));
		}
		if (!((((f & 1) >>> 0) === 0))) {
			throw go$panic(new Go$String("reflect: " + methodName() + " using value obtained using unexported field"));
		}
		if (((f & 4) >>> 0) === 0) {
			throw go$panic(new Go$String("reflect: " + methodName() + " using unaddressable value"));
		}
	};
	go$ptrType(flag).prototype.mustBeAssignable = function() { return new flag(this.go$get()).mustBeAssignable(); };
	Value.Ptr.prototype.Addr = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (((v.flag & 4) >>> 0) === 0) {
			throw go$panic(new Go$String("reflect.Value.Addr of unaddressable value"));
		}
		return new Value.Ptr(v.typ.ptrTo(), v.val, ((((v.flag & 1) >>> 0)) | 352) >>> 0);
	};
	Value.prototype.Addr = function() { return this.go$val.Addr(); };
	Value.Ptr.prototype.Bool = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(1);
		if (!((((v.flag & 2) >>> 0) === 0))) {
			return v.val.go$get();
		}
		return v.val;
	};
	Value.prototype.Bool = function() { return this.go$val.Bool(); };
	Value.Ptr.prototype.Bytes = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(23);
		if (!((v.typ.Elem().Kind() === 8))) {
			throw go$panic(new Go$String("reflect.Value.Bytes of non-byte slice"));
		}
		return v.val.go$get();
	};
	Value.prototype.Bytes = function() { return this.go$val.Bytes(); };
	Value.Ptr.prototype.runes = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(23);
		if (!((v.typ.Elem().Kind() === 5))) {
			throw go$panic(new Go$String("reflect.Value.Bytes of non-rune slice"));
		}
		return v.val.go$get();
	};
	Value.prototype.runes = function() { return this.go$val.runes(); };
	Value.Ptr.prototype.CanAddr = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		return !((((v.flag & 4) >>> 0) === 0));
	};
	Value.prototype.CanAddr = function() { return this.go$val.CanAddr(); };
	Value.Ptr.prototype.CanSet = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		return ((v.flag & 5) >>> 0) === 4;
	};
	Value.prototype.CanSet = function() { return this.go$val.CanSet(); };
	Value.Ptr.prototype.Call = function(in$1) {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(19);
		(new flag(v.flag)).mustBeExported();
		return v.call("Call", in$1);
	};
	Value.prototype.Call = function(in$1) { return this.go$val.Call(in$1); };
	Value.Ptr.prototype.CallSlice = function(in$1) {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(19);
		(new flag(v.flag)).mustBeExported();
		return v.call("CallSlice", in$1);
	};
	Value.prototype.CallSlice = function(in$1) { return this.go$val.CallSlice(in$1); };
	Value.Ptr.prototype.call = function(op, args) {
			var t = this.typ, fn, rcvr;

			if ((this.flag & flagMethod) !== 0) {
				var tuple = methodReceiver(op, this, this.flag >> flagMethodShift);
				t = tuple[0];
				fn = tuple[1];
				rcvr = tuple[2];
			} else {
				fn = this.iword();
			}

			if (fn === go$throwNilPointerError) {
				throw go$panic(new Go$String("reflect.Value.Call: call of nil function"));
			}

			var isSlice = (op === "CallSlice");
			var n = t.NumIn();
			if (isSlice) {
				if (!t.IsVariadic()) {
					throw go$panic(new Go$String("reflect: CallSlice of non-variadic function"));
				}
				if (args.length < n) {
					throw go$panic(new Go$String("reflect: CallSlice with too few input arguments"));
				}
				if (args.length > n) {
					throw go$panic(new Go$String("reflect: CallSlice with too many input arguments"));
				}
			} else {
				if (t.IsVariadic()) {
					n--;
				}
				if (args.length < n) {
					throw go$panic(new Go$String("reflect: Call with too few input arguments"));
				}
				if (!t.IsVariadic() && args.length > n) {
					throw go$panic(new Go$String("reflect: Call with too many input arguments"));
				}
			}
			var i;
			for (i = 0; i < args.length; i++) {
				if (args.array[args.offset + i].Kind() === Invalid) {
					throw go$panic(new Go$String("reflect: " + op + " using zero Value argument"));
				}
			}
			for (i = 0; i < n; i++) {
				var xt = args.array[args.offset + i].Type(), targ = t.In(i);
				if (!xt.AssignableTo(targ)) {
					throw go$panic(new Go$String("reflect: " + op + " using " + xt.String() + " as type " + targ.String()));
				}
			}
			if (!isSlice && t.IsVariadic()) {
				var m = args.length - n;
				var slice = MakeSlice(t.In(n), m, m);
				var elem = t.In(n).Elem();
				for (i = 0; i < m; i++) {
					var x = args.array[args.offset + n + i];
					var xt = x.Type();
					if (!xt.AssignableTo(elem)) {
						throw go$panic(new Go$String("reflect: cannot use " + xt.String() + " as type " + elem.String() + " in " + op));
					}
					slice.Index(i).Set(x);
				}
				args = new (go$sliceType(Value))(go$sliceToArray(args).slice(0, n).concat([slice]));
			}

			if (args.length !== t.NumIn()) {
				throw go$panic(new Go$String("reflect.Value.Call: wrong argument count"));
			}

			var argsArray = new Go$Array(t.NumIn());
			for (i = 0; i < t.NumIn(); i++) {
				argsArray[i] = args.array[args.offset + i].assignTo("reflect.Value.Call", t.In(i), go$ptrType(go$emptyInterface).nil).iword();
			}
			var results = fn.apply(rcvr, argsArray);
			if (t.NumOut() === 0) {
				results = [];
			} else if (t.NumOut() === 1) {
				results = [results];
			}
			for (i = 0; i < t.NumOut(); i++) {
				var typ = t.Out(i);
				var flag = typ.Kind() << flagKindShift;
				results[i] = new Value.Ptr(typ, results[i], flag);
			}
			return new (go$sliceType(Value))(results);
		};
	Value.prototype.call = function() { return this.go$val.call(); };
	methodReceiver = function(op, v, i) {
			var m, t;
			if (v.typ.Kind() === Interface) {
				var tt = v.typ.interfaceType;
				if (i < 0 || i >= tt.methods.length) {
					throw go$panic(new Go$String("reflect: internal error: invalid method index"));
				}
				if (v.IsNil()) {
					throw go$panic(new Go$String("reflect: " + op + " of method on nil interface value"));
				}
				m = tt.methods.array[i];
				t = m.typ;
			} else {
				var ut = v.typ.uncommon();
				if (ut === uncommonType.Ptr.nil || i < 0 || i >= ut.methods.length) {
					throw go$panic(new Go$String("reflect: internal error: invalid method index"));
				}
				m = ut.methods.array[i];
				t = m.mtyp;
			}
			if (m.pkgPath.go$get !== go$throwNilPointerError) {
				throw go$panic(new Go$String("reflect: " + op + " of unexported method"));
			}
			var name = m.name.go$get()
			if (go$reservedKeywords.indexOf(name) !== -1) {
				name += "$";
			}
			var rcvr = v.iword();
			if (isWrapped(v.typ)) {
				rcvr = new v.typ.jsType(rcvr);
			}
			return [t, rcvr[name], rcvr];
		};
	Value.Ptr.prototype.Close = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(18);
		(new flag(v.flag)).mustBeExported();
		chanclose(v.iword());
	};
	Value.prototype.Close = function() { return this.go$val.Close(); };
	Value.Ptr.prototype.Complex = function() {
		var _struct, v, k, _ref, x, x$1;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 15) {
			if (!((((v.flag & 2) >>> 0) === 0))) {
				return (x = v.val.go$get(), new Go$Complex128(x.real, x.imag));
			}
			return (x$1 = v.val, new Go$Complex128(x$1.real, x$1.imag));
		} else if (_ref === 16) {
			return v.val.go$get();
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.Complex", k));
	};
	Value.prototype.Complex = function() { return this.go$val.Complex(); };
	Value.Ptr.prototype.Elem = function() {
			switch (this.kind()) {
			case Interface:
				var val = this.iword();
				if (val === null) {
					return new Value.Ptr();
				}
				if (val.constructor.kind === undefined) { // js.Object
					return new Value.Ptr(Go$String.reflectType(), go$toString(val), String << flagKindShift);
				}
				var typ = val.constructor.reflectType();
				var fl = this.flag & flagRO;
				fl |= typ.Kind() << flagKindShift;
				if (typ.Size() > ptrSize && typ.Kind() !== Array && typ.Kind() !== Struct) {
					return new Value.Ptr(typ, go$newDataPointer(val.go$val, Go$String.Ptr), fl | flagIndir);
				}
				return new Value.Ptr(typ, val.go$val, fl);

			case Ptr:
				var val = this.iword();
				if (this.IsNil()) {
					return new Value.Ptr();
				}
				var tt = this.typ.ptrType;
				var fl = (this.flag & flagRO) | flagIndir | flagAddr;
				fl |= tt.elem.Kind() << flagKindShift;
				return new Value.Ptr(tt.elem, val, fl);
			}
			throw go$panic(new ValueError.Ptr("reflect.Value.Elem", this.kind()));
		};
	Value.prototype.Elem = function() { return this.go$val.Elem(); };
	Value.Ptr.prototype.Field = function(i) {
			this.mustBe(Struct);
			var tt = this.typ.structType;
			if (i < 0 || i >= tt.fields.length) {
				throw go$panic(new Go$String("reflect: Field index out of range"));
			}
			var field = tt.fields.array[i];
			var name = this.typ.jsType.fields[i][0];
			var typ = field.typ;
			var fl = this.flag & (flagRO | flagIndir | flagAddr);
			if (field.pkgPath.go$get !== go$throwNilPointerError) {
				fl |= flagRO;
			}
			fl |= typ.Kind() << flagKindShift;
			if (((this.flag & flagIndir) !== 0 || typ.Size() > ptrSize) && typ.Kind() !== Array && typ.Kind() !== Struct) {
				var struct = this.val;
				return new Value.Ptr(typ, new (go$ptrType(typ.jsType))(function() { return struct[name]; }, function(v) { struct[name] = v; }), fl | flagIndir);
			}
			return new Value.Ptr(typ, this.val[name], fl);
		};
	Value.prototype.Field = function() { return this.go$val.Field(); };
	Value.Ptr.prototype.FieldByIndex = function(index) {
		var _struct, v, _ref, _i, x, i, _struct$1, _struct$2, _struct$3;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(25);
		_ref = index;
		_i = 0;
		while (_i < _ref.length) {
			x = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if (i > 0) {
				if ((v.Kind() === 22) && (v.Elem().Kind() === 25)) {
					v = (_struct$1 = v.Elem(), new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
				}
			}
			v = (_struct$2 = v.Field(x), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
			_i++;
		}
		return (_struct$3 = v, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag));
	};
	Value.prototype.FieldByIndex = function(index) { return this.go$val.FieldByIndex(index); };
	Value.Ptr.prototype.FieldByName = function(name) {
		var _struct, v, _tuple, _struct$1, f, ok, _struct$2;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(25);
		_tuple = v.typ.FieldByName(name); f = (_struct$1 = _tuple[0], new StructField.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, _struct$1.Tag, _struct$1.Offset, _struct$1.Index, _struct$1.Anonymous)); ok = _tuple[1];
		if (ok) {
			return (_struct$2 = v.FieldByIndex(f.Index), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
		}
		return new Value.Ptr((go$ptrType(rtype)).nil, 0, 0);
	};
	Value.prototype.FieldByName = function(name) { return this.go$val.FieldByName(name); };
	Value.Ptr.prototype.FieldByNameFunc = function(match) {
		var _struct, v, _tuple, _struct$1, f, ok, _struct$2;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(25);
		_tuple = v.typ.FieldByNameFunc(match); f = (_struct$1 = _tuple[0], new StructField.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, _struct$1.Tag, _struct$1.Offset, _struct$1.Index, _struct$1.Anonymous)); ok = _tuple[1];
		if (ok) {
			return (_struct$2 = v.FieldByIndex(f.Index), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
		}
		return new Value.Ptr((go$ptrType(rtype)).nil, 0, 0);
	};
	Value.prototype.FieldByNameFunc = function(match) { return this.go$val.FieldByNameFunc(match); };
	Value.Ptr.prototype.Float = function() {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 13) {
			if (!((((v.flag & 2) >>> 0) === 0))) {
				return go$coerceFloat32(v.val.go$get());
			}
			return go$coerceFloat32(v.val);
		} else if (_ref === 14) {
			if (!((((v.flag & 2) >>> 0) === 0))) {
				return v.val.go$get();
			}
			return v.val;
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.Float", k));
	};
	Value.prototype.Float = function() { return this.go$val.Float(); };
	Value.Ptr.prototype.Index = function(i) {
			var k = this.kind();
			switch (k) {
			case Array:
				var tt = this.typ.arrayType;
				if (i < 0 || i >= tt.len) {
					throw go$panic(new Go$String("reflect: array index out of range"));
				}
				var typ = tt.elem;
				var fl = this.flag & (flagRO | flagIndir | flagAddr);
				fl |= typ.Kind() << flagKindShift;
				if (((this.flag & flagIndir) !== 0 || typ.Size() > ptrSize) && typ.Kind() !== Array && typ.Kind() !== Struct) {
					var array = this.val;
					return new Value.Ptr(typ, new (go$ptrType(typ.jsType))(function() { return array[i]; }, function(v) { array[i] = v; }), fl | flagIndir);
				}
				return new Value.Ptr(typ, this.iword()[i], fl);
			case Slice:
				if (i < 0 || i >= this.iword().length) {
					throw go$panic(new Go$String("reflect: slice index out of range"));
				}
				var typ = this.typ.sliceType.elem;
				var fl = flagAddr | flagIndir | (this.flag & flagRO);
				fl |= typ.Kind() << flagKindShift;
				i += this.iword().offset;
				var array = this.iword().array;
				if (typ.Kind() === Struct) {
					return new Value.Ptr(typ, array[i], fl);
				}
				return new Value.Ptr(typ, new (go$ptrType(typ.jsType))(function() { return array[i]; }, function(v) { array[i] = v; }), fl);
			case String:
				var string = this.iword();
				if (i < 0 || i >= string.length) {
					throw go$panic(new Go$String("reflect: string index out of range"));
				}
				var fl = (this.flag & flagRO) | (Uint8 << flagKindShift);
				return new Value.Ptr(uint8Type, string.charCodeAt(i), fl);
			}
			throw go$panic(new ValueError.Ptr("reflect.Value.Index", k));
		};
	Value.prototype.Index = function() { return this.go$val.Index(); };
	Value.Ptr.prototype.Int = function() {
		var _struct, v, k, p, v$1, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		p = 0;
		if (!((((v.flag & 2) >>> 0) === 0))) {
			p = v.val;
		} else {
			p = new (go$ptrType(Go$UnsafePointer))(function() { return v.val; }, function(v$1) { v.val = v$1;; });
		}
		_ref = k;
		if (_ref === 2) {
			return new Go$Int64(0, p.go$get());
		} else if (_ref === 3) {
			return new Go$Int64(0, p.go$get());
		} else if (_ref === 4) {
			return new Go$Int64(0, p.go$get());
		} else if (_ref === 5) {
			return new Go$Int64(0, p.go$get());
		} else if (_ref === 6) {
			return p.go$get();
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.Int", k));
	};
	Value.prototype.Int = function() { return this.go$val.Int(); };
	Value.Ptr.prototype.CanInterface = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (v.flag === 0) {
			throw go$panic(new ValueError.Ptr("reflect.Value.CanInterface", 0));
		}
		return ((v.flag & 1) >>> 0) === 0;
	};
	Value.prototype.CanInterface = function() { return this.go$val.CanInterface(); };
	Value.Ptr.prototype.Interface = function() {
		var i, _struct, v, _struct$1;
		i = null;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		i = valueInterface((_struct$1 = v, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), true);
		return i;
	};
	Value.prototype.Interface = function() { return this.go$val.Interface(); };
	valueInterface = function(v, safe) {
			if (v.flag === 0) {
				throw go$panic(new ValueError.Ptr("reflect.Value.Interface", 0));
			}
			if (safe && (v.flag & flagRO) !== 0) {
				throw go$panic(new Go$String("reflect.Value.Interface: cannot return value obtained from unexported field or method"))
			}
			if ((v.flag & flagMethod) !== 0) {
				v = makeMethodValue("Interface", v);
			}
			if (isWrapped(v.typ)) {
				return new v.typ.jsType(v.iword());
			}
			return v.iword();
		};
	Value.Ptr.prototype.InterfaceData = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(20);
		return go$mapArray(v.val, function(entry) { return entry; });
	};
	Value.prototype.InterfaceData = function() { return this.go$val.InterfaceData(); };
	Value.Ptr.prototype.IsValid = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		return !((v.flag === 0));
	};
	Value.prototype.IsValid = function() { return this.go$val.IsValid(); };
	Value.Ptr.prototype.Kind = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		return (new flag(v.flag)).kind();
	};
	Value.prototype.Kind = function() { return this.go$val.Kind(); };
	Value.Ptr.prototype.MapIndex = function(key) {
		var _struct, v, tt, _struct$1, _tuple, word, ok, typ, fl;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(21);
		tt = v.typ.mapType;
		key = (_struct$1 = key.assignTo("reflect.Value.MapIndex", tt.key, (go$ptrType(go$emptyInterface)).nil), new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
		_tuple = mapaccess(v.typ, v.iword(), key.iword()); word = _tuple[0]; ok = _tuple[1];
		if (!ok) {
			return new Value.Ptr((go$ptrType(rtype)).nil, 0, 0);
		}
		typ = tt.elem;
		fl = ((((v.flag | key.flag) >>> 0)) & 1) >>> 0;
		if (typ.size > 4) {
			fl = (fl | 2) >>> 0;
		}
		fl = (fl | (((typ.Kind() >>> 0) << 4 >>> 0))) >>> 0;
		return new Value.Ptr(typ, word, fl);
	};
	Value.prototype.MapIndex = function(key) { return this.go$val.MapIndex(key); };
	Value.Ptr.prototype.MapKeys = function() {
		var _struct, v, tt, keyType, fl, m, mlen, it, a, i, _tuple, keyWord, ok;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(21);
		tt = v.typ.mapType;
		keyType = tt.key;
		fl = (v.flag & 1) >>> 0;
		fl = (fl | (((keyType.Kind() >>> 0) << 4 >>> 0))) >>> 0;
		if (keyType.size > 4) {
			fl = (fl | 2) >>> 0;
		}
		m = v.iword();
		mlen = 0;
		if (!(m === 0)) {
			mlen = maplen(m);
		}
		it = mapiterinit(v.typ, m);
		a = (go$sliceType(Value)).make(mlen, 0, function() { return new Value.Ptr(); });
		i = 0;
		i = 0;
		while (i < a.length) {
			_tuple = mapiterkey(it); keyWord = _tuple[0]; ok = _tuple[1];
			if (!ok) {
				break;
			}
			(i < 0 || i >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + i] = new Value.Ptr(keyType, keyWord, fl);
			mapiternext(it);
			i = i + 1 >> 0;
		}
		return go$subslice(a, 0, i);
	};
	Value.prototype.MapKeys = function() { return this.go$val.MapKeys(); };
	Value.Ptr.prototype.Method = function(i) {
		var _struct, v, fl;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (v.typ === (go$ptrType(rtype)).nil) {
			throw go$panic(new ValueError.Ptr("reflect.Value.Method", 0));
		}
		if (!((((v.flag & 8) >>> 0) === 0)) || i < 0 || i >= v.typ.NumMethod()) {
			throw go$panic(new Go$String("reflect: Method index out of range"));
		}
		if ((v.typ.Kind() === 20) && v.IsNil()) {
			throw go$panic(new Go$String("reflect: Method on nil interface value"));
		}
		fl = (v.flag & 3) >>> 0;
		fl = (fl | 304) >>> 0;
		fl = (fl | (((((i >>> 0) << 9 >>> 0) | 8) >>> 0))) >>> 0;
		return new Value.Ptr(v.typ, v.val, fl);
	};
	Value.prototype.Method = function(i) { return this.go$val.Method(i); };
	Value.Ptr.prototype.NumMethod = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (v.typ === (go$ptrType(rtype)).nil) {
			throw go$panic(new ValueError.Ptr("reflect.Value.NumMethod", 0));
		}
		if (!((((v.flag & 8) >>> 0) === 0))) {
			return 0;
		}
		return v.typ.NumMethod();
	};
	Value.prototype.NumMethod = function() { return this.go$val.NumMethod(); };
	Value.Ptr.prototype.MethodByName = function(name) {
		var _struct, v, _tuple, _struct$1, _struct$2, m, ok, _struct$3;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (v.typ === (go$ptrType(rtype)).nil) {
			throw go$panic(new ValueError.Ptr("reflect.Value.MethodByName", 0));
		}
		if (!((((v.flag & 8) >>> 0) === 0))) {
			return new Value.Ptr((go$ptrType(rtype)).nil, 0, 0);
		}
		_tuple = v.typ.MethodByName(name); m = (_struct$1 = _tuple[0], new Method.Ptr(_struct$1.Name, _struct$1.PkgPath, _struct$1.Type, (_struct$2 = _struct$1.Func, new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag)), _struct$1.Index)); ok = _tuple[1];
		if (!ok) {
			return new Value.Ptr((go$ptrType(rtype)).nil, 0, 0);
		}
		return (_struct$3 = v.Method(m.Index), new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag));
	};
	Value.prototype.MethodByName = function(name) { return this.go$val.MethodByName(name); };
	Value.Ptr.prototype.NumField = function() {
		var _struct, v, tt;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(25);
		tt = v.typ.structType;
		return tt.fields.length;
	};
	Value.prototype.NumField = function() { return this.go$val.NumField(); };
	Value.Ptr.prototype.OverflowComplex = function(x) {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 15) {
			return overflowFloat32(x.real) || overflowFloat32(x.imag);
		} else if (_ref === 16) {
			return false;
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.OverflowComplex", k));
	};
	Value.prototype.OverflowComplex = function(x) { return this.go$val.OverflowComplex(x); };
	Value.Ptr.prototype.OverflowFloat = function(x) {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 13) {
			return overflowFloat32(x);
		} else if (_ref === 14) {
			return false;
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.OverflowFloat", k));
	};
	Value.prototype.OverflowFloat = function(x) { return this.go$val.OverflowFloat(x); };
	overflowFloat32 = function(x) {
		if (x < 0) {
			x = -x;
		}
		return 3.4028234663852886e+38 < x && x <= 1.7976931348623157e+308;
	};
	Value.Ptr.prototype.OverflowInt = function(x) {
		var _struct, v, k, _ref, x$1, bitSize, trunc;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 2 || _ref === 3 || _ref === 4 || _ref === 5 || _ref === 6) {
			bitSize = (x$1 = v.typ.size, (((x$1 >>> 16 << 16) * 8 >>> 0) + (x$1 << 16 >>> 16) * 8) >>> 0);
			trunc = go$shiftRightInt64((go$shiftLeft64(x, ((64 - bitSize >>> 0)))), ((64 - bitSize >>> 0)));
			return !((x.high === trunc.high && x.low === trunc.low));
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.OverflowInt", k));
	};
	Value.prototype.OverflowInt = function(x) { return this.go$val.OverflowInt(x); };
	Value.Ptr.prototype.OverflowUint = function(x) {
		var _struct, v, k, _ref, x$1, bitSize, trunc;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 7 || _ref === 12 || _ref === 8 || _ref === 9 || _ref === 10 || _ref === 11) {
			bitSize = (x$1 = v.typ.size, (((x$1 >>> 16 << 16) * 8 >>> 0) + (x$1 << 16 >>> 16) * 8) >>> 0);
			trunc = go$shiftRightUint64((go$shiftLeft64(x, ((64 - bitSize >>> 0)))), ((64 - bitSize >>> 0)));
			return !((x.high === trunc.high && x.low === trunc.low));
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.OverflowUint", k));
	};
	Value.prototype.OverflowUint = function(x) { return this.go$val.OverflowUint(x); };
	Value.Ptr.prototype.Recv = function() {
		var x, ok, _struct, v, _tuple, _struct$1, _struct$2;
		x = new Value.Ptr();
		ok = false;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(18);
		(new flag(v.flag)).mustBeExported();
		_tuple = v.recv(false); x = (_struct$1 = _tuple[0], new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)); ok = _tuple[1];
		return [(_struct$2 = x, new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag)), ok];
	};
	Value.prototype.Recv = function() { return this.go$val.Recv(); };
	Value.Ptr.prototype.recv = function(nb) {
		var val, ok, _struct, v, tt, _tuple, word, selected, typ, fl, _struct$1;
		val = new Value.Ptr();
		ok = false;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		tt = v.typ.chanType;
		if (((tt.dir >> 0) & 1) === 0) {
			throw go$panic(new Go$String("reflect: recv on send-only channel"));
		}
		_tuple = chanrecv(v.typ, v.iword(), nb); word = _tuple[0]; selected = _tuple[1]; ok = _tuple[2];
		if (selected) {
			typ = tt.elem;
			fl = (typ.Kind() >>> 0) << 4 >>> 0;
			if (typ.size > 4) {
				fl = (fl | 2) >>> 0;
			}
			val = new Value.Ptr(typ, word, fl);
		}
		return [(_struct$1 = val, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), ok];
	};
	Value.prototype.recv = function(nb) { return this.go$val.recv(nb); };
	Value.Ptr.prototype.Send = function(x) {
		var _struct, v, _struct$1;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(18);
		(new flag(v.flag)).mustBeExported();
		v.send((_struct$1 = x, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), false);
	};
	Value.prototype.Send = function(x) { return this.go$val.Send(x); };
	Value.Ptr.prototype.send = function(x, nb) {
		var selected, _struct, v, tt, _struct$1;
		selected = false;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		tt = v.typ.chanType;
		if (((tt.dir >> 0) & 2) === 0) {
			throw go$panic(new Go$String("reflect: send on recv-only channel"));
		}
		(new flag(x.flag)).mustBeExported();
		x = (_struct$1 = x.assignTo("reflect.Value.Send", tt.elem, (go$ptrType(go$emptyInterface)).nil), new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
		selected = chansend(v.typ, v.iword(), x.iword(), nb);
		return selected;
	};
	Value.prototype.send = function(x, nb) { return this.go$val.send(x, nb); };
	Value.Ptr.prototype.Set = function(x) {
			this.mustBeAssignable();
			x.mustBeExported();
			if ((this.flag & flagIndir) !== 0) {
				switch (this.typ.Kind()) {
				case Array:
					go$copyArray(this.val, x.val);
					return;
				case Interface:
					this.val.go$set(valueInterface(x, false));
					return;
				case Struct:
					copyStruct(this.val, x.val, this.typ);
					return;
				default:
					this.val.go$set(x.iword());
					return;
				}
			}
			this.val = x.val;
		};
	Value.prototype.Set = function() { return this.go$val.Set(); };
	Value.Ptr.prototype.SetBool = function(x) {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		(new flag(v.flag)).mustBe(1);
		v.val.go$set(x);
	};
	Value.prototype.SetBool = function(x) { return this.go$val.SetBool(x); };
	Value.Ptr.prototype.SetBytes = function(x) {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		(new flag(v.flag)).mustBe(23);
		if (!((v.typ.Elem().Kind() === 8))) {
			throw go$panic(new Go$String("reflect.Value.SetBytes of non-byte slice"));
		}
		v.val.go$set(x);
	};
	Value.prototype.SetBytes = function(x) { return this.go$val.SetBytes(x); };
	Value.Ptr.prototype.SetComplex = function(x) {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 15) {
			v.val.go$set(new Go$Complex64(x.real, x.imag));
		} else if (_ref === 16) {
			v.val.go$set(x);
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.SetComplex", k));
		}
	};
	Value.prototype.SetComplex = function(x) { return this.go$val.SetComplex(x); };
	Value.Ptr.prototype.SetFloat = function(x) {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 13) {
			v.val.go$set(x);
		} else if (_ref === 14) {
			v.val.go$set(x);
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.SetFloat", k));
		}
	};
	Value.prototype.SetFloat = function(x) { return this.go$val.SetFloat(x); };
	Value.Ptr.prototype.SetInt = function(x) {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 2) {
			v.val.go$set(((x.low + ((x.high >> 31) * 4294967296)) >> 0));
		} else if (_ref === 3) {
			v.val.go$set(((x.low + ((x.high >> 31) * 4294967296)) << 24 >> 24));
		} else if (_ref === 4) {
			v.val.go$set(((x.low + ((x.high >> 31) * 4294967296)) << 16 >> 16));
		} else if (_ref === 5) {
			v.val.go$set(((x.low + ((x.high >> 31) * 4294967296)) >> 0));
		} else if (_ref === 6) {
			v.val.go$set(x);
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.SetInt", k));
		}
	};
	Value.prototype.SetInt = function(x) { return this.go$val.SetInt(x); };
	Value.Ptr.prototype.SetLen = function(n) {
			this.mustBeAssignable();
			this.mustBe(Slice);
			var s = this.val.go$get();
			if (n < 0 || n > s.capacity) {
				throw go$panic(new Go$String("reflect: slice length out of range in SetLen"));
			}
			var newSlice = new this.typ.jsType(s.array);
			newSlice.offset = s.offset;
			newSlice.length = n;
			newSlice.capacity = s.capacity;
			this.val.go$set(newSlice);
		};
	Value.prototype.SetLen = function() { return this.go$val.SetLen(); };
	Value.Ptr.prototype.SetCap = function(n) {
			this.mustBeAssignable();
			this.mustBe(Slice);
			var s = this.val.go$get();
			if (n < s.length || n > s.capacity) {
				throw go$panic(new Go$String("reflect: slice capacity out of range in SetCap"));
			}
			var newSlice = new this.typ.jsType(s.array);
			newSlice.offset = s.offset;
			newSlice.length = s.length;
			newSlice.capacity = n;
			this.val.go$set(newSlice);
		};
	Value.prototype.SetCap = function() { return this.go$val.SetCap(); };
	Value.Ptr.prototype.SetMapIndex = function(key, val) {
		var _struct, v, tt, _struct$1, _struct$2;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(21);
		(new flag(v.flag)).mustBeExported();
		(new flag(key.flag)).mustBeExported();
		tt = v.typ.mapType;
		key = (_struct$1 = key.assignTo("reflect.Value.SetMapIndex", tt.key, (go$ptrType(go$emptyInterface)).nil), new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
		if (!(val.typ === (go$ptrType(rtype)).nil)) {
			(new flag(val.flag)).mustBeExported();
			val = (_struct$2 = val.assignTo("reflect.Value.SetMapIndex", tt.elem, (go$ptrType(go$emptyInterface)).nil), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
		}
		mapassign(v.typ, v.iword(), key.iword(), val.iword(), !(val.typ === (go$ptrType(rtype)).nil));
	};
	Value.prototype.SetMapIndex = function(key, val) { return this.go$val.SetMapIndex(key, val); };
	Value.Ptr.prototype.SetUint = function(x) {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 7) {
			v.val.go$set((x.low >>> 0));
		} else if (_ref === 8) {
			v.val.go$set((x.low << 24 >>> 24));
		} else if (_ref === 9) {
			v.val.go$set((x.low << 16 >>> 16));
		} else if (_ref === 10) {
			v.val.go$set((x.low >>> 0));
		} else if (_ref === 11) {
			v.val.go$set(x);
		} else if (_ref === 12) {
			v.val.go$set((x.low >>> 0));
		} else {
			throw go$panic(new ValueError.Ptr("reflect.Value.SetUint", k));
		}
	};
	Value.prototype.SetUint = function(x) { return this.go$val.SetUint(x); };
	Value.Ptr.prototype.SetPointer = function(x) {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		(new flag(v.flag)).mustBe(26);
		v.val.go$set(x);
	};
	Value.prototype.SetPointer = function(x) { return this.go$val.SetPointer(x); };
	Value.Ptr.prototype.SetString = function(x) {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBeAssignable();
		(new flag(v.flag)).mustBe(24);
		v.val.go$set(x);
	};
	Value.prototype.SetString = function(x) { return this.go$val.SetString(x); };
	Value.Ptr.prototype.Slice = function(i, j) {
			var typ, s, cap;
			var kind = this.kind();
			switch (kind) {
			case Array:
				if ((this.flag & flagAddr) === 0) {
					throw go$panic(new Go$String("reflect.Value.Slice: slice of unaddressable array"));
				}
				var tt = this.typ.arrayType;
				cap = tt.len;
				typ = SliceOf(tt.elem);
				s = new typ.jsType(this.iword());
				break;
			case Slice:
				typ = this.typ.sliceType;
				s = this.iword();
				cap = s.capacity;
				break;
			case String:
				s = this.iword();
				if (i < 0 || j < i || j > s.length) {
					throw go$panic(new Go$String("reflect.Value.Slice: string slice index out of bounds"));
				}
				return new Value.Ptr(this.typ, go$newDataPointer(s.substring(i, j), Go$String.Ptr), this.flag);
			default:
				throw go$panic(new ValueError.Ptr("reflect.Value.Slice", kind));
			}

			if (i < 0 || j < i || j > cap) {
				throw go$panic(new Go$String("reflect.Value.Slice: slice index out of bounds"));
			}

			var fl = (this.flag & flagRO) | flagIndir | (Slice << flagKindShift);
			return new Value.Ptr(typ.common(), go$newDataPointer(go$subslice(s, i, j), typ.ptrTo().jsType), fl);
		};
	Value.prototype.Slice = function() { return this.go$val.Slice(); };
	Value.Ptr.prototype.Slice3 = function(i, j, k) {
			var typ, s, cap;
			var kind = this.kind();
			switch (kind) {
			case Array:
				if ((this.flag & flagAddr) === 0) {
					throw go$panic(new Go$String("reflect.Value.Slice3: slice of unaddressable array"));
				}
				var tt = this.typ.arrayType;
				cap = tt.len;
				typ = SliceOf(tt.elem);
				s = new typ.jsType(this.iword());
				break;
			case Slice:
				typ = this.typ.sliceType;
				s = this.iword();
				cap = s.capacity;
				break;
			default:
				throw go$panic(new ValueError.Ptr("reflect.Value.Slice3", kind));
			}

			if (i < 0 || j < i || k < j || k > cap) {
				throw go$panic(new Go$String("reflect.Value.Slice3: slice index out of bounds"));
			}

			var fl = (this.flag & flagRO) | flagIndir | (Slice << flagKindShift);
			return new Value.Ptr(typ.common(), go$newDataPointer(go$subslice(s, i, j, k), typ.ptrTo().jsType), fl);
		};
	Value.prototype.Slice3 = function() { return this.go$val.Slice3(); };
	Value.Ptr.prototype.String = function() {
		var _struct, v, k, _ref;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		_ref = k;
		if (_ref === 0) {
			return "<invalid Value>";
		} else if (_ref === 24) {
			return v.val.go$get();
		}
		return "<" + v.typ.String() + " Value>";
	};
	Value.prototype.String = function() { return this.go$val.String(); };
	Value.Ptr.prototype.TryRecv = function() {
		var x, ok, _struct, v, _tuple, _struct$1, _struct$2;
		x = new Value.Ptr();
		ok = false;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(18);
		(new flag(v.flag)).mustBeExported();
		_tuple = v.recv(true); x = (_struct$1 = _tuple[0], new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)); ok = _tuple[1];
		return [(_struct$2 = x, new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag)), ok];
	};
	Value.prototype.TryRecv = function() { return this.go$val.TryRecv(); };
	Value.Ptr.prototype.TrySend = function(x) {
		var _struct, v, _struct$1;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		(new flag(v.flag)).mustBe(18);
		(new flag(v.flag)).mustBeExported();
		return v.send((_struct$1 = x, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), true);
	};
	Value.prototype.TrySend = function(x) { return this.go$val.TrySend(x); };
	Value.Ptr.prototype.Type = function() {
		var _struct, v, f, i, tt, x, m, ut, x$1, m$1;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		f = v.flag;
		if (f === 0) {
			throw go$panic(new ValueError.Ptr("reflect.Value.Type", 0));
		}
		if (((f & 8) >>> 0) === 0) {
			return v.typ;
		}
		i = (v.flag >> 0) >> 9 >> 0;
		if (v.typ.Kind() === 20) {
			tt = v.typ.interfaceType;
			if (i < 0 || i >= tt.methods.length) {
				throw go$panic(new Go$String("reflect: internal error: invalid method index"));
			}
			m = (x = tt.methods, ((i < 0 || i >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + i]));
			return m.typ;
		}
		ut = v.typ.uncommonType.uncommon();
		if (ut === (go$ptrType(uncommonType)).nil || i < 0 || i >= ut.methods.length) {
			throw go$panic(new Go$String("reflect: internal error: invalid method index"));
		}
		m$1 = (x$1 = ut.methods, ((i < 0 || i >= x$1.length) ? go$throwRuntimeError("index out of range") : x$1.array[x$1.offset + i]));
		return m$1.mtyp;
	};
	Value.prototype.Type = function() { return this.go$val.Type(); };
	Value.Ptr.prototype.Uint = function() {
		var _struct, v, k, p, v$1, _ref, x;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		k = (new flag(v.flag)).kind();
		p = 0;
		if (!((((v.flag & 2) >>> 0) === 0))) {
			p = v.val;
		} else {
			p = new (go$ptrType(Go$UnsafePointer))(function() { return v.val; }, function(v$1) { v.val = v$1;; });
		}
		_ref = k;
		if (_ref === 7) {
			return new Go$Uint64(0, p.go$get());
		} else if (_ref === 8) {
			return new Go$Uint64(0, p.go$get());
		} else if (_ref === 9) {
			return new Go$Uint64(0, p.go$get());
		} else if (_ref === 10) {
			return new Go$Uint64(0, p.go$get());
		} else if (_ref === 11) {
			return p.go$get();
		} else if (_ref === 12) {
			return (x = p.go$get(), new Go$Uint64(0, x.constructor === Number ? x : 1));
		}
		throw go$panic(new ValueError.Ptr("reflect.Value.Uint", k));
	};
	Value.prototype.Uint = function() { return this.go$val.Uint(); };
	Value.Ptr.prototype.UnsafeAddr = function() {
		var _struct, v;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (v.typ === (go$ptrType(rtype)).nil) {
			throw go$panic(new ValueError.Ptr("reflect.Value.UnsafeAddr", 0));
		}
		if (((v.flag & 4) >>> 0) === 0) {
			throw go$panic(new Go$String("reflect.Value.UnsafeAddr of unaddressable value"));
		}
		return v.val;
	};
	Value.prototype.UnsafeAddr = function() { return this.go$val.UnsafeAddr(); };
	typesMustMatch = function(what, t1, t2) {
		if (!(go$interfaceIsEqual(t1, t2))) {
			throw go$panic(new Go$String(what + ": " + t1.String() + " != " + t2.String()));
		}
	};
	New = go$pkg.New = function(typ) {
		var ptr, fl;
		if (go$interfaceIsEqual(typ, null)) {
			throw go$panic(new Go$String("reflect: New(nil)"));
		}
		ptr = unsafe_New((typ !== null && typ.constructor === (go$ptrType(rtype)) ? typ.go$val : go$typeAssertionFailed(typ, (go$ptrType(rtype)))));
		fl = 352;
		return new Value.Ptr(typ.common().ptrTo(), ptr, fl);
	};
	Value.Ptr.prototype.assignTo = function(context, dst, target) {
		var _struct, v, _struct$1, _struct$2, fl, _struct$3, x;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (!((((v.flag & 8) >>> 0) === 0))) {
			v = (_struct$2 = makeMethodValue(context, (_struct$1 = v, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag))), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
		}
		if (directlyAssignable(dst, v.typ)) {
			v.typ = dst;
			fl = (v.flag & 7) >>> 0;
			fl = (fl | (((dst.Kind() >>> 0) << 4 >>> 0))) >>> 0;
			return new Value.Ptr(dst, v.val, fl);
		} else if (implements$1(dst, v.typ)) {
			if (target === (go$ptrType(go$emptyInterface)).nil) {
				target = go$newDataPointer(null, (go$ptrType(go$emptyInterface)));
			}
			x = valueInterface((_struct$3 = v, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), false);
			if (dst.NumMethod() === 0) {
				target.go$set(x);
			} else {
				ifaceE2I(dst, x, target);
			}
			return new Value.Ptr(dst, target, 322);
		}
		throw go$panic(new Go$String(context + ": value of type " + v.typ.String() + " is not assignable to type " + dst.String()));
	};
	Value.prototype.assignTo = function(context, dst, target) { return this.go$val.assignTo(context, dst, target); };
	Value.Ptr.prototype.Convert = function(t) {
		var _struct, v, _struct$1, _struct$2, op, _struct$3, _struct$4;
		v = (_struct = this, new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if (!((((v.flag & 8) >>> 0) === 0))) {
			v = (_struct$2 = makeMethodValue("Convert", (_struct$1 = v, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag))), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
		}
		op = convertOp(t.common(), v.typ);
		if (op === go$throwNilPointerError) {
			throw go$panic(new Go$String("reflect.Value.Convert: value of type " + v.typ.String() + " cannot be converted to type " + t.String()));
		}
		return (_struct$4 = op((_struct$3 = v, new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag)), t), new Value.Ptr(_struct$4.typ, _struct$4.val, _struct$4.flag));
	};
	Value.prototype.Convert = function(t) { return this.go$val.Convert(t); };
	convertOp = function(dst, src) {
		var _ref, _ref$1, _ref$2, _ref$3, _ref$4, _ref$5, _ref$6;
		_ref = src.Kind();
		if (_ref === 2 || _ref === 3 || _ref === 4 || _ref === 5 || _ref === 6) {
			_ref$1 = dst.Kind();
			if (_ref$1 === 2 || _ref$1 === 3 || _ref$1 === 4 || _ref$1 === 5 || _ref$1 === 6 || _ref$1 === 7 || _ref$1 === 8 || _ref$1 === 9 || _ref$1 === 10 || _ref$1 === 11 || _ref$1 === 12) {
				return cvtInt;
			} else if (_ref$1 === 13 || _ref$1 === 14) {
				return cvtIntFloat;
			} else if (_ref$1 === 24) {
				return cvtIntString;
			}
		} else if (_ref === 7 || _ref === 8 || _ref === 9 || _ref === 10 || _ref === 11 || _ref === 12) {
			_ref$2 = dst.Kind();
			if (_ref$2 === 2 || _ref$2 === 3 || _ref$2 === 4 || _ref$2 === 5 || _ref$2 === 6 || _ref$2 === 7 || _ref$2 === 8 || _ref$2 === 9 || _ref$2 === 10 || _ref$2 === 11 || _ref$2 === 12) {
				return cvtUint;
			} else if (_ref$2 === 13 || _ref$2 === 14) {
				return cvtUintFloat;
			} else if (_ref$2 === 24) {
				return cvtUintString;
			}
		} else if (_ref === 13 || _ref === 14) {
			_ref$3 = dst.Kind();
			if (_ref$3 === 2 || _ref$3 === 3 || _ref$3 === 4 || _ref$3 === 5 || _ref$3 === 6) {
				return cvtFloatInt;
			} else if (_ref$3 === 7 || _ref$3 === 8 || _ref$3 === 9 || _ref$3 === 10 || _ref$3 === 11 || _ref$3 === 12) {
				return cvtFloatUint;
			} else if (_ref$3 === 13 || _ref$3 === 14) {
				return cvtFloat;
			}
		} else if (_ref === 15 || _ref === 16) {
			_ref$4 = dst.Kind();
			if (_ref$4 === 15 || _ref$4 === 16) {
				return cvtComplex;
			}
		} else if (_ref === 24) {
			if ((dst.Kind() === 23) && dst.Elem().PkgPath() === "") {
				_ref$5 = dst.Elem().Kind();
				if (_ref$5 === 8) {
					return cvtStringBytes;
				} else if (_ref$5 === 5) {
					return cvtStringRunes;
				}
			}
		} else if (_ref === 23) {
			if ((dst.Kind() === 24) && src.Elem().PkgPath() === "") {
				_ref$6 = src.Elem().Kind();
				if (_ref$6 === 8) {
					return cvtBytesString;
				} else if (_ref$6 === 5) {
					return cvtRunesString;
				}
			}
		}
		if (haveIdenticalUnderlyingType(dst, src)) {
			return cvtDirect;
		}
		if ((dst.Kind() === 22) && dst.Name() === "" && (src.Kind() === 22) && src.Name() === "" && haveIdenticalUnderlyingType(dst.Elem().common(), src.Elem().common())) {
			return cvtDirect;
		}
		if (implements$1(dst, src)) {
			if (src.Kind() === 20) {
				return cvtI2I;
			}
			return cvtT2I;
		}
		return go$throwNilPointerError;
	};
	makeInt = function(f, bits, typ) {
			var val;
			switch (typ.Kind()) {
			case Int8:
				val = bits.low << 24 >> 24;
				break;
			case Int16:
				val = bits.low << 16 >> 16;
				break;
			case Int:
			case Int32:
				val = bits.low >> 0;
				break;
			case Int64:
				return new Value.Ptr(typ, go$newDataPointer(new Go$Int64(bits.high, bits.low), typ.ptrTo().jsType), f | flagIndir | (Int64 << flagKindShift));
			case Uint8:
				val = bits.low << 24 >>> 24;
				break;
			case Uint16:
				val = bits.low << 16 >>> 16;
				break;
			case Uint64:
				return new Value.Ptr(typ, go$newDataPointer(bits, typ.ptrTo().jsType), f | flagIndir | (Int64 << flagKindShift));
			case Uint:
			case Uint32:
			case Uintptr:
				val = bits.low >>> 0;
				break;
			}
			return new Value.Ptr(typ, val, f | (typ.Kind() << flagKindShift));
		};
	makeFloat = function(f, v, t) {
		var typ, ptr, w, _ref, v$1, v$2;
		typ = t.common();
		if (typ.size > 4) {
			ptr = unsafe_New(typ);
			ptr.go$set(v);
			return new Value.Ptr(typ, ptr, (((f | 2) >>> 0) | ((typ.Kind() >>> 0) << 4 >>> 0)) >>> 0);
		}
		w = 0;
		_ref = typ.size;
		if (_ref === 4) {
			new (go$ptrType(iword))(function() { return w; }, function(v$1) { w = v$1;; }).go$set(v);
		} else if (_ref === 8) {
			new (go$ptrType(iword))(function() { return w; }, function(v$2) { w = v$2;; }).go$set(v);
		}
		return new Value.Ptr(typ, w, (f | ((typ.Kind() >>> 0) << 4 >>> 0)) >>> 0);
	};
	makeComplex = function(f, v, typ) {
			return new Value.Ptr(typ, new typ.jsType(v.real, v.imag), f | (typ.Kind() << flagKindShift));
		};
	makeString = function(f, v, t) {
		var _struct, ret, _struct$1;
		ret = (_struct = New(t).Elem(), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		ret.SetString(v);
		ret.flag = ((ret.flag & ~4) | f) >>> 0;
		return (_struct$1 = ret, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
	};
	cvtInt = function(v, t) {
		var x, _struct;
		return (_struct = makeInt((v.flag & 1) >>> 0, (x = v.Int(), new Go$Uint64(x.high, x.low)), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtUint = function(v, t) {
		var _struct;
		return (_struct = makeInt((v.flag & 1) >>> 0, v.Uint(), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtFloatInt = function(v, t) {
		var x, _struct;
		return (_struct = makeInt((v.flag & 1) >>> 0, (x = new Go$Int64(0, v.Float()), new Go$Uint64(x.high, x.low)), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtFloatUint = function(v, t) {
		var _struct;
		return (_struct = makeInt((v.flag & 1) >>> 0, new Go$Uint64(0, v.Float()), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtIntFloat = function(v, t) {
		var _struct;
		return (_struct = makeFloat((v.flag & 1) >>> 0, go$flatten64(v.Int()), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtUintFloat = function(v, t) {
		var _struct;
		return (_struct = makeFloat((v.flag & 1) >>> 0, go$flatten64(v.Uint()), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtFloat = function(v, t) {
		var _struct;
		return (_struct = makeFloat((v.flag & 1) >>> 0, v.Float(), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtComplex = function(v, t) {
		var _struct;
		return (_struct = makeComplex((v.flag & 1) >>> 0, v.Complex(), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtIntString = function(v, t) {
		var _struct;
		return (_struct = makeString((v.flag & 1) >>> 0, go$encodeRune(v.Int().low), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtUintString = function(v, t) {
		var _struct;
		return (_struct = makeString((v.flag & 1) >>> 0, go$encodeRune(v.Uint().low), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtBytesString = function(v, t) {
		var _struct;
		return (_struct = makeString((v.flag & 1) >>> 0, go$bytesToString(v.Bytes()), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtStringBytes = function(v, typ) {
			return new Value.Ptr(typ, new typ.jsType(go$stringToBytes(v.iword())), (v.flag & flagRO) | (Slice << flagKindShift));
		};
	cvtRunesString = function(v, t) {
		var _struct;
		return (_struct = makeString((v.flag & 1) >>> 0, go$runesToString(v.runes()), t), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
	};
	cvtStringRunes = function(v, typ) {
			return new Value.Ptr(typ, new typ.jsType(go$stringToRunes(v.iword())), (v.flag & flagRO) | (Slice << flagKindShift));
		};
	cvtDirect = function(v, typ) {
			var srcVal = v.iword();
			if (srcVal === v.typ.jsType.nil) {
				if (typ.Kind() === Slice) {
					return new Value.Ptr(typ, go$newDataPointer(typ.jsType.nil, typ.ptrTo().jsType), v.flag);
				}
				return new Value.Ptr(typ, typ.jsType.nil, v.flag);
			}

			var val;
			switch (typ.Kind()) {
			case Chan:
				val = new typ.jsType();
				break;
			case Slice:
				var slice = new typ.jsType(srcVal.array);
				slice.length = srcVal.length;
				slice.cap = srcVal.cap;
				val = go$newDataPointer(slice, typ.ptrTo().jsType);
				break;
			case Ptr:
				if (typ.Elem().Kind() === Struct) {
					if (typ.Elem() === v.typ.Elem()) {
						val = srcVal;
					}
					val = new typ.jsType();
					copyStruct(val, srcVal, typ.Elem());
					break;
				}
				val = new typ.jsType(srcVal.go$get, srcVal.go$set);
				break;
			case Struct:
				val = new typ.jsType.Ptr();
				copyStruct(val, srcVal, typ);
				break;
			case Array:
			case Func:
			case Interface:
			case Map:
			case String:
				val = v.val;
				break;
			default:
				throw go$panic(new ValueError.Ptr("reflect.Convert", typ.Kind()));
			}
			return new Value.Ptr(typ, val, (v.flag & (flagRO | flagIndir)) | (typ.Kind() << flagKindShift));
		};
	cvtT2I = function(v, typ) {
		var target, _struct, x;
		target = go$newDataPointer(null, (go$ptrType(go$emptyInterface)));
		x = valueInterface((_struct = v, new Value.Ptr(_struct.typ, _struct.val, _struct.flag)), false);
		if (typ.NumMethod() === 0) {
			target.go$set(x);
		} else {
			ifaceE2I((typ !== null && typ.constructor === (go$ptrType(rtype)) ? typ.go$val : go$typeAssertionFailed(typ, (go$ptrType(rtype)))), x, target);
		}
		return new Value.Ptr(typ.common(), target, (((((v.flag & 1) >>> 0) | 2) >>> 0) | 320) >>> 0);
	};
	cvtI2I = function(v, typ) {
		var _struct, ret, _struct$1, _struct$2, _struct$3;
		if (v.IsNil()) {
			ret = (_struct = Zero(typ), new Value.Ptr(_struct.typ, _struct.val, _struct.flag));
			ret.flag = (ret.flag | (((v.flag & 1) >>> 0))) >>> 0;
			return (_struct$1 = ret, new Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
		}
		return (_struct$3 = cvtT2I((_struct$2 = v.Elem(), new Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag)), typ), new Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag));
	};
	call = function() {
		throw go$panic("Native function not implemented: call");
	};
	ifaceE2I = function(t, src, dst) {
			dst.go$set(src);
		};

			var Invalid = 0;
			var Bool = 1;
			var Int = 2;
			var Int8 = 3;
			var Int16 = 4;
			var Int32 = 5;
			var Int64 = 6;
			var Uint = 7;
			var Uint8 = 8;
			var Uint16 = 9;
			var Uint32 = 10;
			var Uint64 = 11;
			var Uintptr = 12;
			var Float32 = 13;
			var Float64 = 14;
			var Complex64 = 15;
			var Complex128 = 16;
			var Array = 17;
			var Chan = 18;
			var Func = 19;
			var Interface = 20;
			var Map = 21;
			var Ptr = 22;
			var Slice = 23;
			var String = 24;
			var Struct = 25;
			var UnsafePointer = 26;

			var RecvDir = 1;
			var SendDir = 2;
			var BothDir = 3;

			var flagRO = 1;
			var flagIndir = 2;
			var flagAddr = 4;
			var flagMethod = 8;
			var flagKindShift = 4;
			var flagKindWidth = 5;
			var flagKindMask = 31;
			var flagMethodShift = 9;

			var ptrSize = 4;

			go$reflect = {
				rtype: rtype.Ptr, uncommonType: uncommonType.Ptr, method: method.Ptr, arrayType: arrayType.Ptr, chanType: chanType.Ptr, funcType: funcType.Ptr, interfaceType: interfaceType.Ptr, mapType: mapType.Ptr, ptrType: ptrType.Ptr, sliceType: sliceType.Ptr, structType: structType.Ptr,
				imethod: imethod.Ptr, structField: structField.Ptr,
				kinds: { Bool: Bool, Int: Int, Int8: Int8, Int16: Int16, Int32: Int32, Int64: Int64, Uint: Uint, Uint8: Uint8, Uint16: Uint16, Uint32: Uint32, Uint64: Uint64, Uintptr: Uintptr, Float32: Float32, Float64: Float64, Complex64: Complex64, Complex128: Complex128, Array: Array, Chan: Chan, Func: Func, Interface: Interface, Map: Map, Ptr: Ptr, Slice: Slice, String: String, Struct: Struct, UnsafePointer: UnsafePointer },
				RecvDir: RecvDir, SendDir: SendDir, BothDir: BothDir
			};

			var deepValueEqual = function(v1, v2, visited) {
				if (!v1.IsValid() || !v2.IsValid()) {
					return !v1.IsValid() && !v2.IsValid();
				}
				if (v1.Type() !== v2.Type()) {
					return false;
				}

				var i;
				switch(v1.Kind()) {
				case Array:
				case Map:
				case Slice:
				case Struct:
					for (i = 0; i < visited.length; i++) {
						var entry = visited[i];
						if (v1.val === entry[0] && v2.val === entry[1]) {
							return true;
						}
					}
					visited.push([v1.val, v2.val]);
				}

				switch(v1.Kind()) {
				case Array:
				case Slice:
					if (v1.Kind() === Slice) {
						if (v1.IsNil() !== v2.IsNil()) {
							return false;
						}
						if (v1.iword() === v2.iword()) {
							return true;
						}
					}
					var n = v1.Len();
					if (n !== v2.Len()) {
						return false;
					}
					for (i = 0; i < n; i++) {
						if (!deepValueEqual(v1.Index(i), v2.Index(i), visited)) {
							return false;
						}
					}
					return true;
				case Interface:
					if (v1.IsNil() || v2.IsNil()) {
						return v1.IsNil() && v2.IsNil();
					}
					return deepValueEqual(v1.Elem(), v2.Elem(), visited);
				case Ptr:
					return deepValueEqual(v1.Elem(), v2.Elem(), visited);
				case Struct:
					var n = v1.NumField();
					for (i = 0; i < n; i++) {
						if (!deepValueEqual(v1.Field(i), v2.Field(i), visited)) {
							return false;
						}
					}
					return true;
				case Map:
					if (v1.IsNil() !== v2.IsNil()) {
						return false;
					}
					if (v1.iword() === v2.iword()) {
						return true;
					}
					var keys = v1.MapKeys();
					if (keys.length !== v2.Len()) {
						return false;
					}
					for (i = 0; i < keys.length; i++) {
						var k = keys.array[i];
						if (!deepValueEqual(v1.MapIndex(k), v2.MapIndex(k), visited)) {
							return false;
						}
					}
					return true;
				case Func:
					return v1.IsNil() && v2.IsNil();
				}

				return go$interfaceIsEqual(valueInterface(v1, false), valueInterface(v2, false));
			};
			go$pkg.init = function() {
		mapIter.init([["t", "t", "reflect", Type, ""], ["m", "m", "reflect", js.Object, ""], ["keys", "keys", "reflect", js.Object, ""], ["i", "i", "reflect", Go$Int, ""]]);
		Type.init([["Align", "", (go$funcType([], [Go$Int], false))], ["AssignableTo", "", (go$funcType([Type], [Go$Bool], false))], ["Bits", "", (go$funcType([], [Go$Int], false))], ["ChanDir", "", (go$funcType([], [ChanDir], false))], ["ConvertibleTo", "", (go$funcType([Type], [Go$Bool], false))], ["Elem", "", (go$funcType([], [Type], false))], ["Field", "", (go$funcType([Go$Int], [StructField], false))], ["FieldAlign", "", (go$funcType([], [Go$Int], false))], ["FieldByIndex", "", (go$funcType([(go$sliceType(Go$Int))], [StructField], false))], ["FieldByName", "", (go$funcType([Go$String], [StructField, Go$Bool], false))], ["FieldByNameFunc", "", (go$funcType([(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false))], ["Implements", "", (go$funcType([Type], [Go$Bool], false))], ["In", "", (go$funcType([Go$Int], [Type], false))], ["IsVariadic", "", (go$funcType([], [Go$Bool], false))], ["Key", "", (go$funcType([], [Type], false))], ["Kind", "", (go$funcType([], [Kind], false))], ["Len", "", (go$funcType([], [Go$Int], false))], ["Method", "", (go$funcType([Go$Int], [Method], false))], ["MethodByName", "", (go$funcType([Go$String], [Method, Go$Bool], false))], ["Name", "", (go$funcType([], [Go$String], false))], ["NumField", "", (go$funcType([], [Go$Int], false))], ["NumIn", "", (go$funcType([], [Go$Int], false))], ["NumMethod", "", (go$funcType([], [Go$Int], false))], ["NumOut", "", (go$funcType([], [Go$Int], false))], ["Out", "", (go$funcType([Go$Int], [Type], false))], ["PkgPath", "", (go$funcType([], [Go$String], false))], ["Size", "", (go$funcType([], [Go$Uintptr], false))], ["String", "", (go$funcType([], [Go$String], false))], ["common", "reflect", (go$funcType([], [(go$ptrType(rtype))], false))], ["uncommon", "reflect", (go$funcType([], [(go$ptrType(uncommonType))], false))]]);
		Kind.methods = [["String", "", [], [Go$String], false, -1]];
		(go$ptrType(Kind)).methods = [["String", "", [], [Go$String], false, -1]];
		rtype.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 9]];
		(go$ptrType(rtype)).methods = [["Align", "", [], [Go$Int], false, -1], ["AssignableTo", "", [Type], [Go$Bool], false, -1], ["Bits", "", [], [Go$Int], false, -1], ["ChanDir", "", [], [ChanDir], false, -1], ["ConvertibleTo", "", [Type], [Go$Bool], false, -1], ["Elem", "", [], [Type], false, -1], ["Field", "", [Go$Int], [StructField], false, -1], ["FieldAlign", "", [], [Go$Int], false, -1], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, -1], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, -1], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, -1], ["Implements", "", [Type], [Go$Bool], false, -1], ["In", "", [Go$Int], [Type], false, -1], ["IsVariadic", "", [], [Go$Bool], false, -1], ["Key", "", [], [Type], false, -1], ["Kind", "", [], [Kind], false, -1], ["Len", "", [], [Go$Int], false, -1], ["Method", "", [Go$Int], [Method], false, -1], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, -1], ["Name", "", [], [Go$String], false, -1], ["NumField", "", [], [Go$Int], false, -1], ["NumIn", "", [], [Go$Int], false, -1], ["NumMethod", "", [], [Go$Int], false, -1], ["NumOut", "", [], [Go$Int], false, -1], ["Out", "", [Go$Int], [Type], false, -1], ["PkgPath", "", [], [Go$String], false, -1], ["Size", "", [], [Go$Uintptr], false, -1], ["String", "", [], [Go$String], false, -1], ["common", "reflect", [], [(go$ptrType(rtype))], false, -1], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, -1], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 9]];
		rtype.init([["size", "size", "reflect", Go$Uintptr, ""], ["hash", "hash", "reflect", Go$Uint32, ""], ["_$2", "_", "reflect", Go$Uint8, ""], ["align", "align", "reflect", Go$Uint8, ""], ["fieldAlign", "fieldAlign", "reflect", Go$Uint8, ""], ["kind", "kind", "reflect", Go$Uint8, ""], ["alg", "alg", "reflect", (go$ptrType(Go$Uintptr)), ""], ["gc", "gc", "reflect", Go$UnsafePointer, ""], ["string", "string", "reflect", (go$ptrType(Go$String)), ""], ["uncommonType", "", "reflect", (go$ptrType(uncommonType)), ""], ["ptrToThis", "ptrToThis", "reflect", (go$ptrType(rtype)), ""]]);
		method.init([["name", "name", "reflect", (go$ptrType(Go$String)), ""], ["pkgPath", "pkgPath", "reflect", (go$ptrType(Go$String)), ""], ["mtyp", "mtyp", "reflect", (go$ptrType(rtype)), ""], ["typ", "typ", "reflect", (go$ptrType(rtype)), ""], ["ifn", "ifn", "reflect", Go$UnsafePointer, ""], ["tfn", "tfn", "reflect", Go$UnsafePointer, ""]]);
		(go$ptrType(uncommonType)).methods = [["Method", "", [Go$Int], [Method], false, -1], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, -1], ["Name", "", [], [Go$String], false, -1], ["NumMethod", "", [], [Go$Int], false, -1], ["PkgPath", "", [], [Go$String], false, -1], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, -1]];
		uncommonType.init([["name", "name", "reflect", (go$ptrType(Go$String)), ""], ["pkgPath", "pkgPath", "reflect", (go$ptrType(Go$String)), ""], ["methods", "methods", "reflect", (go$sliceType(method)), ""]]);
		ChanDir.methods = [["String", "", [], [Go$String], false, -1]];
		(go$ptrType(ChanDir)).methods = [["String", "", [], [Go$String], false, -1]];
		arrayType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(arrayType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		arrayType.init([["rtype", "", "reflect", rtype, "reflect:\"array\""], ["elem", "elem", "reflect", (go$ptrType(rtype)), ""], ["slice", "slice", "reflect", (go$ptrType(rtype)), ""], ["len", "len", "reflect", Go$Uintptr, ""]]);
		chanType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(chanType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		chanType.init([["rtype", "", "reflect", rtype, "reflect:\"chan\""], ["elem", "elem", "reflect", (go$ptrType(rtype)), ""], ["dir", "dir", "reflect", Go$Uintptr, ""]]);
		funcType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(funcType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		funcType.init([["rtype", "", "reflect", rtype, "reflect:\"func\""], ["dotdotdot", "dotdotdot", "reflect", Go$Bool, ""], ["in$2", "in", "reflect", (go$sliceType((go$ptrType(rtype)))), ""], ["out", "out", "reflect", (go$sliceType((go$ptrType(rtype)))), ""]]);
		imethod.init([["name", "name", "reflect", (go$ptrType(Go$String)), ""], ["pkgPath", "pkgPath", "reflect", (go$ptrType(Go$String)), ""], ["typ", "typ", "reflect", (go$ptrType(rtype)), ""]]);
		interfaceType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(interfaceType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, -1], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, -1], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, -1], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		interfaceType.init([["rtype", "", "reflect", rtype, "reflect:\"interface\""], ["methods", "methods", "reflect", (go$sliceType(imethod)), ""]]);
		mapType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(mapType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		mapType.init([["rtype", "", "reflect", rtype, "reflect:\"map\""], ["key", "key", "reflect", (go$ptrType(rtype)), ""], ["elem", "elem", "reflect", (go$ptrType(rtype)), ""], ["bucket", "bucket", "reflect", (go$ptrType(rtype)), ""], ["hmap", "hmap", "reflect", (go$ptrType(rtype)), ""]]);
		ptrType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(ptrType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		ptrType.init([["rtype", "", "reflect", rtype, "reflect:\"ptr\""], ["elem", "elem", "reflect", (go$ptrType(rtype)), ""]]);
		sliceType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(sliceType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, 0], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, 0], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, 0], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, 0], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		sliceType.init([["rtype", "", "reflect", rtype, "reflect:\"slice\""], ["elem", "elem", "reflect", (go$ptrType(rtype)), ""]]);
		structField.init([["name", "name", "reflect", (go$ptrType(Go$String)), ""], ["pkgPath", "pkgPath", "reflect", (go$ptrType(Go$String)), ""], ["typ", "typ", "reflect", (go$ptrType(rtype)), ""], ["tag", "tag", "reflect", (go$ptrType(Go$String)), ""], ["offset", "offset", "reflect", Go$Uintptr, ""]]);
		structType.methods = [["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		(go$ptrType(structType)).methods = [["Align", "", [], [Go$Int], false, 0], ["AssignableTo", "", [Type], [Go$Bool], false, 0], ["Bits", "", [], [Go$Int], false, 0], ["ChanDir", "", [], [ChanDir], false, 0], ["ConvertibleTo", "", [Type], [Go$Bool], false, 0], ["Elem", "", [], [Type], false, 0], ["Field", "", [Go$Int], [StructField], false, -1], ["FieldAlign", "", [], [Go$Int], false, 0], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [StructField], false, -1], ["FieldByName", "", [Go$String], [StructField, Go$Bool], false, -1], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [StructField, Go$Bool], false, -1], ["Implements", "", [Type], [Go$Bool], false, 0], ["In", "", [Go$Int], [Type], false, 0], ["IsVariadic", "", [], [Go$Bool], false, 0], ["Key", "", [], [Type], false, 0], ["Kind", "", [], [Kind], false, 0], ["Len", "", [], [Go$Int], false, 0], ["Method", "", [Go$Int], [Method], false, 0], ["MethodByName", "", [Go$String], [Method, Go$Bool], false, 0], ["Name", "", [], [Go$String], false, 0], ["NumField", "", [], [Go$Int], false, 0], ["NumIn", "", [], [Go$Int], false, 0], ["NumMethod", "", [], [Go$Int], false, 0], ["NumOut", "", [], [Go$Int], false, 0], ["Out", "", [Go$Int], [Type], false, 0], ["PkgPath", "", [], [Go$String], false, 0], ["Size", "", [], [Go$Uintptr], false, 0], ["String", "", [], [Go$String], false, 0], ["common", "reflect", [], [(go$ptrType(rtype))], false, 0], ["ptrTo", "reflect", [], [(go$ptrType(rtype))], false, 0], ["uncommon", "reflect", [], [(go$ptrType(uncommonType))], false, 0]];
		structType.init([["rtype", "", "reflect", rtype, "reflect:\"struct\""], ["fields", "fields", "reflect", (go$sliceType(structField)), ""]]);
		Method.init([["Name", "Name", "", Go$String, ""], ["PkgPath", "PkgPath", "", Go$String, ""], ["Type", "Type", "", Type, ""], ["Func", "Func", "", Value, ""], ["Index", "Index", "", Go$Int, ""]]);
		StructField.init([["Name", "Name", "", Go$String, ""], ["PkgPath", "PkgPath", "", Go$String, ""], ["Type", "Type", "", Type, ""], ["Tag", "Tag", "", StructTag, ""], ["Offset", "Offset", "", Go$Uintptr, ""], ["Index", "Index", "", (go$sliceType(Go$Int)), ""], ["Anonymous", "Anonymous", "", Go$Bool, ""]]);
		StructTag.methods = [["Get", "", [Go$String], [Go$String], false, -1]];
		(go$ptrType(StructTag)).methods = [["Get", "", [Go$String], [Go$String], false, -1]];
		fieldScan.init([["typ", "typ", "reflect", (go$ptrType(structType)), ""], ["index", "index", "reflect", (go$sliceType(Go$Int)), ""]]);
		Value.methods = [["Addr", "", [], [Value], false, -1], ["Bool", "", [], [Go$Bool], false, -1], ["Bytes", "", [], [(go$sliceType(Go$Uint8))], false, -1], ["Call", "", [(go$sliceType(Value))], [(go$sliceType(Value))], false, -1], ["CallSlice", "", [(go$sliceType(Value))], [(go$sliceType(Value))], false, -1], ["CanAddr", "", [], [Go$Bool], false, -1], ["CanInterface", "", [], [Go$Bool], false, -1], ["CanSet", "", [], [Go$Bool], false, -1], ["Cap", "", [], [Go$Int], false, -1], ["Close", "", [], [], false, -1], ["Complex", "", [], [Go$Complex128], false, -1], ["Convert", "", [Type], [Value], false, -1], ["Elem", "", [], [Value], false, -1], ["Field", "", [Go$Int], [Value], false, -1], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [Value], false, -1], ["FieldByName", "", [Go$String], [Value], false, -1], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [Value], false, -1], ["Float", "", [], [Go$Float64], false, -1], ["Index", "", [Go$Int], [Value], false, -1], ["Int", "", [], [Go$Int64], false, -1], ["Interface", "", [], [go$emptyInterface], false, -1], ["InterfaceData", "", [], [(go$arrayType(Go$Uintptr, 2))], false, -1], ["IsNil", "", [], [Go$Bool], false, -1], ["IsValid", "", [], [Go$Bool], false, -1], ["Kind", "", [], [Kind], false, -1], ["Len", "", [], [Go$Int], false, -1], ["MapIndex", "", [Value], [Value], false, -1], ["MapKeys", "", [], [(go$sliceType(Value))], false, -1], ["Method", "", [Go$Int], [Value], false, -1], ["MethodByName", "", [Go$String], [Value], false, -1], ["NumField", "", [], [Go$Int], false, -1], ["NumMethod", "", [], [Go$Int], false, -1], ["OverflowComplex", "", [Go$Complex128], [Go$Bool], false, -1], ["OverflowFloat", "", [Go$Float64], [Go$Bool], false, -1], ["OverflowInt", "", [Go$Int64], [Go$Bool], false, -1], ["OverflowUint", "", [Go$Uint64], [Go$Bool], false, -1], ["Pointer", "", [], [Go$Uintptr], false, -1], ["Recv", "", [], [Value, Go$Bool], false, -1], ["Send", "", [Value], [], false, -1], ["Set", "", [Value], [], false, -1], ["SetBool", "", [Go$Bool], [], false, -1], ["SetBytes", "", [(go$sliceType(Go$Uint8))], [], false, -1], ["SetCap", "", [Go$Int], [], false, -1], ["SetComplex", "", [Go$Complex128], [], false, -1], ["SetFloat", "", [Go$Float64], [], false, -1], ["SetInt", "", [Go$Int64], [], false, -1], ["SetLen", "", [Go$Int], [], false, -1], ["SetMapIndex", "", [Value, Value], [], false, -1], ["SetPointer", "", [Go$UnsafePointer], [], false, -1], ["SetString", "", [Go$String], [], false, -1], ["SetUint", "", [Go$Uint64], [], false, -1], ["Slice", "", [Go$Int, Go$Int], [Value], false, -1], ["Slice3", "", [Go$Int, Go$Int, Go$Int], [Value], false, -1], ["String", "", [], [Go$String], false, -1], ["TryRecv", "", [], [Value, Go$Bool], false, -1], ["TrySend", "", [Value], [Go$Bool], false, -1], ["Type", "", [], [Type], false, -1], ["Uint", "", [], [Go$Uint64], false, -1], ["UnsafeAddr", "", [], [Go$Uintptr], false, -1], ["assignTo", "reflect", [Go$String, (go$ptrType(rtype)), (go$ptrType(go$emptyInterface))], [Value], false, -1], ["call", "reflect", [Go$String, (go$sliceType(Value))], [(go$sliceType(Value))], false, -1], ["iword", "reflect", [], [iword], false, -1], ["kind", "reflect", [], [Kind], false, 2], ["mustBe", "reflect", [Kind], [], false, 2], ["mustBeAssignable", "reflect", [], [], false, 2], ["mustBeExported", "reflect", [], [], false, 2], ["recv", "reflect", [Go$Bool], [Value, Go$Bool], false, -1], ["runes", "reflect", [], [(go$sliceType(Go$Int32))], false, -1], ["send", "reflect", [Value, Go$Bool], [Go$Bool], false, -1], ["setRunes", "reflect", [(go$sliceType(Go$Int32))], [], false, -1]];
		(go$ptrType(Value)).methods = [["Addr", "", [], [Value], false, -1], ["Bool", "", [], [Go$Bool], false, -1], ["Bytes", "", [], [(go$sliceType(Go$Uint8))], false, -1], ["Call", "", [(go$sliceType(Value))], [(go$sliceType(Value))], false, -1], ["CallSlice", "", [(go$sliceType(Value))], [(go$sliceType(Value))], false, -1], ["CanAddr", "", [], [Go$Bool], false, -1], ["CanInterface", "", [], [Go$Bool], false, -1], ["CanSet", "", [], [Go$Bool], false, -1], ["Cap", "", [], [Go$Int], false, -1], ["Close", "", [], [], false, -1], ["Complex", "", [], [Go$Complex128], false, -1], ["Convert", "", [Type], [Value], false, -1], ["Elem", "", [], [Value], false, -1], ["Field", "", [Go$Int], [Value], false, -1], ["FieldByIndex", "", [(go$sliceType(Go$Int))], [Value], false, -1], ["FieldByName", "", [Go$String], [Value], false, -1], ["FieldByNameFunc", "", [(go$funcType([Go$String], [Go$Bool], false))], [Value], false, -1], ["Float", "", [], [Go$Float64], false, -1], ["Index", "", [Go$Int], [Value], false, -1], ["Int", "", [], [Go$Int64], false, -1], ["Interface", "", [], [go$emptyInterface], false, -1], ["InterfaceData", "", [], [(go$arrayType(Go$Uintptr, 2))], false, -1], ["IsNil", "", [], [Go$Bool], false, -1], ["IsValid", "", [], [Go$Bool], false, -1], ["Kind", "", [], [Kind], false, -1], ["Len", "", [], [Go$Int], false, -1], ["MapIndex", "", [Value], [Value], false, -1], ["MapKeys", "", [], [(go$sliceType(Value))], false, -1], ["Method", "", [Go$Int], [Value], false, -1], ["MethodByName", "", [Go$String], [Value], false, -1], ["NumField", "", [], [Go$Int], false, -1], ["NumMethod", "", [], [Go$Int], false, -1], ["OverflowComplex", "", [Go$Complex128], [Go$Bool], false, -1], ["OverflowFloat", "", [Go$Float64], [Go$Bool], false, -1], ["OverflowInt", "", [Go$Int64], [Go$Bool], false, -1], ["OverflowUint", "", [Go$Uint64], [Go$Bool], false, -1], ["Pointer", "", [], [Go$Uintptr], false, -1], ["Recv", "", [], [Value, Go$Bool], false, -1], ["Send", "", [Value], [], false, -1], ["Set", "", [Value], [], false, -1], ["SetBool", "", [Go$Bool], [], false, -1], ["SetBytes", "", [(go$sliceType(Go$Uint8))], [], false, -1], ["SetCap", "", [Go$Int], [], false, -1], ["SetComplex", "", [Go$Complex128], [], false, -1], ["SetFloat", "", [Go$Float64], [], false, -1], ["SetInt", "", [Go$Int64], [], false, -1], ["SetLen", "", [Go$Int], [], false, -1], ["SetMapIndex", "", [Value, Value], [], false, -1], ["SetPointer", "", [Go$UnsafePointer], [], false, -1], ["SetString", "", [Go$String], [], false, -1], ["SetUint", "", [Go$Uint64], [], false, -1], ["Slice", "", [Go$Int, Go$Int], [Value], false, -1], ["Slice3", "", [Go$Int, Go$Int, Go$Int], [Value], false, -1], ["String", "", [], [Go$String], false, -1], ["TryRecv", "", [], [Value, Go$Bool], false, -1], ["TrySend", "", [Value], [Go$Bool], false, -1], ["Type", "", [], [Type], false, -1], ["Uint", "", [], [Go$Uint64], false, -1], ["UnsafeAddr", "", [], [Go$Uintptr], false, -1], ["assignTo", "reflect", [Go$String, (go$ptrType(rtype)), (go$ptrType(go$emptyInterface))], [Value], false, -1], ["call", "reflect", [Go$String, (go$sliceType(Value))], [(go$sliceType(Value))], false, -1], ["iword", "reflect", [], [iword], false, -1], ["kind", "reflect", [], [Kind], false, 2], ["mustBe", "reflect", [Kind], [], false, 2], ["mustBeAssignable", "reflect", [], [], false, 2], ["mustBeExported", "reflect", [], [], false, 2], ["recv", "reflect", [Go$Bool], [Value, Go$Bool], false, -1], ["runes", "reflect", [], [(go$sliceType(Go$Int32))], false, -1], ["send", "reflect", [Value, Go$Bool], [Go$Bool], false, -1], ["setRunes", "reflect", [(go$sliceType(Go$Int32))], [], false, -1]];
		Value.init([["typ", "typ", "reflect", (go$ptrType(rtype)), ""], ["val", "val", "reflect", Go$UnsafePointer, ""], ["flag", "", "reflect", flag, ""]]);
		flag.methods = [["kind", "reflect", [], [Kind], false, -1], ["mustBe", "reflect", [Kind], [], false, -1], ["mustBeAssignable", "reflect", [], [], false, -1], ["mustBeExported", "reflect", [], [], false, -1]];
		(go$ptrType(flag)).methods = [["kind", "reflect", [], [Kind], false, -1], ["mustBe", "reflect", [Kind], [], false, -1], ["mustBeAssignable", "reflect", [], [], false, -1], ["mustBeExported", "reflect", [], [], false, -1]];
		(go$ptrType(ValueError)).methods = [["Error", "", [], [Go$String], false, -1]];
		ValueError.init([["Method", "Method", "", Go$String, ""], ["Kind", "Kind", "", Kind, ""]]);
		kindNames = new (go$sliceType(Go$String))(["invalid", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128", "array", "chan", "func", "interface", "map", "ptr", "slice", "string", "struct", "unsafe.Pointer"]);
		var x;
		uint8Type = (x = TypeOf(new Go$Uint8(0)), (x !== null && x.constructor === (go$ptrType(rtype)) ? x.go$val : go$typeAssertionFailed(x, (go$ptrType(rtype)))));
		var a;
		a = false;
		if (a) {
			isWrapped(null);
			copyStruct(null, null, null);
		}
	}
	return go$pkg;
})();
go$packages["fmt"] = (function() {
	var go$pkg = {}, strconv = go$packages["strconv"], utf8 = go$packages["unicode/utf8"], errors = go$packages["errors"], io = go$packages["io"], os = go$packages["os"], reflect = go$packages["reflect"], sync = go$packages["sync"], math = go$packages["math"], fmt, State, Formatter, Stringer, GoStringer, buffer, pp, cache, runeUnreader, scanError, ss, ssave, doPrec, newCache, newPrinter, Sprintf, getField, parsenum, intFromArg, parseArgNumber, isSpace, notSpace, indexRune, padZeroBytes, padSpaceBytes, trueBytes, falseBytes, commaSpaceBytes, nilAngleBytes, nilParenBytes, nilBytes, mapBytes, percentBangBytes, missingBytes, badIndexBytes, panicBytes, extraBytes, irparenBytes, bytesBytes, badWidthBytes, badPrecBytes, noVerbBytes, ppFree, intBits, uintptrBits, space, ssFree, complexError, boolError;
	fmt = go$pkg.fmt = go$newType(0, "Struct", "fmt.fmt", "fmt", "fmt", function(intbuf_, buf_, wid_, prec_, widPresent_, precPresent_, minus_, plus_, sharp_, space_, unicode_, uniQuote_, zero_) {
		this.go$val = this;
		this.intbuf = intbuf_ !== undefined ? intbuf_ : go$makeNativeArray("Uint8", 65, function() { return 0; });
		this.buf = buf_ !== undefined ? buf_ : (go$ptrType(buffer)).nil;
		this.wid = wid_ !== undefined ? wid_ : 0;
		this.prec = prec_ !== undefined ? prec_ : 0;
		this.widPresent = widPresent_ !== undefined ? widPresent_ : false;
		this.precPresent = precPresent_ !== undefined ? precPresent_ : false;
		this.minus = minus_ !== undefined ? minus_ : false;
		this.plus = plus_ !== undefined ? plus_ : false;
		this.sharp = sharp_ !== undefined ? sharp_ : false;
		this.space = space_ !== undefined ? space_ : false;
		this.unicode = unicode_ !== undefined ? unicode_ : false;
		this.uniQuote = uniQuote_ !== undefined ? uniQuote_ : false;
		this.zero = zero_ !== undefined ? zero_ : false;
	});
	State = go$pkg.State = go$newType(8, "Interface", "fmt.State", "State", "fmt", null);
	Formatter = go$pkg.Formatter = go$newType(8, "Interface", "fmt.Formatter", "Formatter", "fmt", null);
	Stringer = go$pkg.Stringer = go$newType(8, "Interface", "fmt.Stringer", "Stringer", "fmt", null);
	GoStringer = go$pkg.GoStringer = go$newType(8, "Interface", "fmt.GoStringer", "GoStringer", "fmt", null);
	buffer = go$pkg.buffer = go$newType(12, "Slice", "fmt.buffer", "buffer", "fmt", null);
	pp = go$pkg.pp = go$newType(0, "Struct", "fmt.pp", "pp", "fmt", function(n_, panicking_, erroring_, buf_, arg_, value_, reordered_, goodArgNum_, runeBuf_, fmt_) {
		this.go$val = this;
		this.n = n_ !== undefined ? n_ : 0;
		this.panicking = panicking_ !== undefined ? panicking_ : false;
		this.erroring = erroring_ !== undefined ? erroring_ : false;
		this.buf = buf_ !== undefined ? buf_ : buffer.nil;
		this.arg = arg_ !== undefined ? arg_ : null;
		this.value = value_ !== undefined ? value_ : new reflect.Value.Ptr();
		this.reordered = reordered_ !== undefined ? reordered_ : false;
		this.goodArgNum = goodArgNum_ !== undefined ? goodArgNum_ : false;
		this.runeBuf = runeBuf_ !== undefined ? runeBuf_ : go$makeNativeArray("Uint8", 4, function() { return 0; });
		this.fmt = fmt_ !== undefined ? fmt_ : new fmt.Ptr();
	});
	cache = go$pkg.cache = go$newType(0, "Struct", "fmt.cache", "cache", "fmt", function(mu_, saved_, new$2_) {
		this.go$val = this;
		this.mu = mu_ !== undefined ? mu_ : new sync.Mutex.Ptr();
		this.saved = saved_ !== undefined ? saved_ : (go$sliceType(go$emptyInterface)).nil;
		this.new$2 = new$2_ !== undefined ? new$2_ : go$throwNilPointerError;
	});
	runeUnreader = go$pkg.runeUnreader = go$newType(8, "Interface", "fmt.runeUnreader", "runeUnreader", "fmt", null);
	scanError = go$pkg.scanError = go$newType(0, "Struct", "fmt.scanError", "scanError", "fmt", function(err_) {
		this.go$val = this;
		this.err = err_ !== undefined ? err_ : null;
	});
	ss = go$pkg.ss = go$newType(0, "Struct", "fmt.ss", "ss", "fmt", function(rr_, buf_, peekRune_, prevRune_, count_, atEOF_, ssave_) {
		this.go$val = this;
		this.rr = rr_ !== undefined ? rr_ : null;
		this.buf = buf_ !== undefined ? buf_ : buffer.nil;
		this.peekRune = peekRune_ !== undefined ? peekRune_ : 0;
		this.prevRune = prevRune_ !== undefined ? prevRune_ : 0;
		this.count = count_ !== undefined ? count_ : 0;
		this.atEOF = atEOF_ !== undefined ? atEOF_ : false;
		this.ssave = ssave_ !== undefined ? ssave_ : new ssave.Ptr();
	});
	ssave = go$pkg.ssave = go$newType(0, "Struct", "fmt.ssave", "ssave", "fmt", function(validSave_, nlIsEnd_, nlIsSpace_, argLimit_, limit_, maxWid_) {
		this.go$val = this;
		this.validSave = validSave_ !== undefined ? validSave_ : false;
		this.nlIsEnd = nlIsEnd_ !== undefined ? nlIsEnd_ : false;
		this.nlIsSpace = nlIsSpace_ !== undefined ? nlIsSpace_ : false;
		this.argLimit = argLimit_ !== undefined ? argLimit_ : 0;
		this.limit = limit_ !== undefined ? limit_ : 0;
		this.maxWid = maxWid_ !== undefined ? maxWid_ : 0;
	});
	fmt.Ptr.prototype.clearflags = function() {
		var f;
		f = this;
		f.wid = 0;
		f.widPresent = false;
		f.prec = 0;
		f.precPresent = false;
		f.minus = false;
		f.plus = false;
		f.sharp = false;
		f.space = false;
		f.unicode = false;
		f.uniQuote = false;
		f.zero = false;
	};
	fmt.prototype.clearflags = function() { return this.go$val.clearflags(); };
	fmt.Ptr.prototype.init = function(buf) {
		var f;
		f = this;
		f.buf = buf;
		f.clearflags();
	};
	fmt.prototype.init = function(buf) { return this.go$val.init(buf); };
	fmt.Ptr.prototype.computePadding = function(width) {
		var padding, leftWidth, rightWidth, f, left, w, _tmp, _tmp$1, _tmp$2, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, _tmp$8;
		padding = (go$sliceType(Go$Uint8)).nil;
		leftWidth = 0;
		rightWidth = 0;
		f = this;
		left = !f.minus;
		w = f.wid;
		if (w < 0) {
			left = false;
			w = -w;
		}
		w = w - (width) >> 0;
		if (w > 0) {
			if (left && f.zero) {
				_tmp = padZeroBytes; _tmp$1 = w; _tmp$2 = 0; padding = _tmp; leftWidth = _tmp$1; rightWidth = _tmp$2;
				return [padding, leftWidth, rightWidth];
			}
			if (left) {
				_tmp$3 = padSpaceBytes; _tmp$4 = w; _tmp$5 = 0; padding = _tmp$3; leftWidth = _tmp$4; rightWidth = _tmp$5;
				return [padding, leftWidth, rightWidth];
			} else {
				_tmp$6 = padSpaceBytes; _tmp$7 = 0; _tmp$8 = w; padding = _tmp$6; leftWidth = _tmp$7; rightWidth = _tmp$8;
				return [padding, leftWidth, rightWidth];
			}
		}
		return [padding, leftWidth, rightWidth];
	};
	fmt.prototype.computePadding = function(width) { return this.go$val.computePadding(width); };
	fmt.Ptr.prototype.writePadding = function(n, padding) {
		var f, m;
		f = this;
		while (n > 0) {
			m = n;
			if (m > 65) {
				m = 65;
			}
			f.buf.Write(go$subslice(padding, 0, m));
			n = n - (m) >> 0;
		}
	};
	fmt.prototype.writePadding = function(n, padding) { return this.go$val.writePadding(n, padding); };
	fmt.Ptr.prototype.pad = function(b) {
		var f, _tuple, padding, left, right;
		f = this;
		if (!f.widPresent || (f.wid === 0)) {
			f.buf.Write(b);
			return;
		}
		_tuple = f.computePadding(b.length); padding = _tuple[0]; left = _tuple[1]; right = _tuple[2];
		if (left > 0) {
			f.writePadding(left, padding);
		}
		f.buf.Write(b);
		if (right > 0) {
			f.writePadding(right, padding);
		}
	};
	fmt.prototype.pad = function(b) { return this.go$val.pad(b); };
	fmt.Ptr.prototype.padString = function(s) {
		var f, _tuple, padding, left, right;
		f = this;
		if (!f.widPresent || (f.wid === 0)) {
			f.buf.WriteString(s);
			return;
		}
		_tuple = f.computePadding(utf8.RuneCountInString(s)); padding = _tuple[0]; left = _tuple[1]; right = _tuple[2];
		if (left > 0) {
			f.writePadding(left, padding);
		}
		f.buf.WriteString(s);
		if (right > 0) {
			f.writePadding(right, padding);
		}
	};
	fmt.prototype.padString = function(s) { return this.go$val.padString(s); };
	fmt.Ptr.prototype.fmt_boolean = function(v) {
		var f;
		f = this;
		if (v) {
			f.pad(trueBytes);
		} else {
			f.pad(falseBytes);
		}
	};
	fmt.prototype.fmt_boolean = function(v) { return this.go$val.fmt_boolean(v); };
	fmt.Ptr.prototype.integer = function(a, base, signedness, digits) {
		var f, buf, negative, prec, i, ua, _ref, runeWidth, width, j;
		f = this;
		if (f.precPresent && (f.prec === 0) && (a.high === 0 && a.low === 0)) {
			return;
		}
		buf = go$subslice(new (go$sliceType(Go$Uint8))(f.intbuf), 0);
		if (f.widPresent && f.wid > 65) {
			buf = (go$sliceType(Go$Uint8)).make(f.wid, 0, function() { return 0; });
		}
		negative = signedness === true && (a.high < 0 || (a.high === 0 && a.low < 0));
		if (negative) {
			a = new Go$Int64(-a.high, -a.low);
		}
		prec = 0;
		if (f.precPresent) {
			prec = f.prec;
			f.zero = false;
		} else if (f.zero && f.widPresent && !f.minus && f.wid > 0) {
			prec = f.wid;
			if (negative || f.plus || f.space) {
				prec = prec - 1 >> 0;
			}
		}
		i = buf.length;
		ua = new Go$Uint64(a.high, a.low);
		while ((ua.high > base.high || (ua.high === base.high && ua.low >= base.low))) {
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = digits.charCodeAt(go$flatten64(go$div64(ua, base, true)));
			ua = go$div64(ua, (base), false);
		}
		i = i - 1 >> 0;
		(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = digits.charCodeAt(go$flatten64(ua));
		while (i > 0 && prec > (buf.length - i >> 0)) {
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 48;
		}
		if (f.sharp) {
			_ref = base;
			if ((_ref.high === 0 && _ref.low === 8)) {
				if (!((((i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i]) === 48))) {
					i = i - 1 >> 0;
					(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 48;
				}
			} else if ((_ref.high === 0 && _ref.low === 16)) {
				i = i - 1 >> 0;
				(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = (120 + digits.charCodeAt(10) << 24 >>> 24) - 97 << 24 >>> 24;
				i = i - 1 >> 0;
				(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 48;
			}
		}
		if (f.unicode) {
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 43;
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 85;
		}
		if (negative) {
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 45;
		} else if (f.plus) {
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 43;
		} else if (f.space) {
			i = i - 1 >> 0;
			(i < 0 || i >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + i] = 32;
		}
		if (f.unicode && f.uniQuote && (a.high > 0 || (a.high === 0 && a.low >= 0)) && (a.high < 0 || (a.high === 0 && a.low <= 1114111)) && strconv.IsPrint(((a.low + ((a.high >> 31) * 4294967296)) >> 0))) {
			runeWidth = utf8.RuneLen(((a.low + ((a.high >> 31) * 4294967296)) >> 0));
			width = (2 + runeWidth >> 0) + 1 >> 0;
			go$copySlice(go$subslice(buf, (i - width >> 0)), go$subslice(buf, i));
			i = i - (width) >> 0;
			j = buf.length - width >> 0;
			(j < 0 || j >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + j] = 32;
			j = j + 1 >> 0;
			(j < 0 || j >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + j] = 39;
			j = j + 1 >> 0;
			utf8.EncodeRune(go$subslice(buf, j), ((a.low + ((a.high >> 31) * 4294967296)) >> 0));
			j = j + (runeWidth) >> 0;
			(j < 0 || j >= buf.length) ? go$throwRuntimeError("index out of range") : buf.array[buf.offset + j] = 39;
		}
		f.pad(go$subslice(buf, i));
	};
	fmt.prototype.integer = function(a, base, signedness, digits) { return this.go$val.integer(a, base, signedness, digits); };
	fmt.Ptr.prototype.truncate = function(s) {
		var f, n, _ref, _i, _rune, i;
		f = this;
		if (f.precPresent && f.prec < utf8.RuneCountInString(s)) {
			n = f.prec;
			_ref = s;
			_i = 0;
			while (_i < _ref.length) {
				_rune = go$decodeRune(_ref, _i);
				i = _i;
				if (n === 0) {
					s = s.substring(0, i);
					break;
				}
				n = n - 1 >> 0;
				_i += _rune[1];
			}
		}
		return s;
	};
	fmt.prototype.truncate = function(s) { return this.go$val.truncate(s); };
	fmt.Ptr.prototype.fmt_s = function(s) {
		var f;
		f = this;
		s = f.truncate(s);
		f.padString(s);
	};
	fmt.prototype.fmt_s = function(s) { return this.go$val.fmt_s(s); };
	fmt.Ptr.prototype.fmt_sbx = function(s, b, digits) {
		var f, n, x, buf, i, c;
		f = this;
		n = b.length;
		if (b === (go$sliceType(Go$Uint8)).nil) {
			n = s.length;
		}
		x = (digits.charCodeAt(10) - 97 << 24 >>> 24) + 120 << 24 >>> 24;
		buf = (go$sliceType(Go$Uint8)).nil;
		i = 0;
		while (i < n) {
			if (i > 0 && f.space) {
				buf = go$append(buf, 32);
			}
			if (f.sharp) {
				buf = go$append(buf, 48, x);
			}
			c = 0;
			if (b === (go$sliceType(Go$Uint8)).nil) {
				c = s.charCodeAt(i);
			} else {
				c = ((i < 0 || i >= b.length) ? go$throwRuntimeError("index out of range") : b.array[b.offset + i]);
			}
			buf = go$append(buf, digits.charCodeAt((c >>> 4 << 24 >>> 24)), digits.charCodeAt(((c & 15) >>> 0)));
			i = i + 1 >> 0;
		}
		f.pad(buf);
	};
	fmt.prototype.fmt_sbx = function(s, b, digits) { return this.go$val.fmt_sbx(s, b, digits); };
	fmt.Ptr.prototype.fmt_sx = function(s, digits) {
		var f;
		f = this;
		f.fmt_sbx(s, (go$sliceType(Go$Uint8)).nil, digits);
	};
	fmt.prototype.fmt_sx = function(s, digits) { return this.go$val.fmt_sx(s, digits); };
	fmt.Ptr.prototype.fmt_bx = function(b, digits) {
		var f;
		f = this;
		f.fmt_sbx("", b, digits);
	};
	fmt.prototype.fmt_bx = function(b, digits) { return this.go$val.fmt_bx(b, digits); };
	fmt.Ptr.prototype.fmt_q = function(s) {
		var f, quoted;
		f = this;
		s = f.truncate(s);
		quoted = "";
		if (f.sharp && strconv.CanBackquote(s)) {
			quoted = "`" + s + "`";
		} else {
			if (f.plus) {
				quoted = strconv.QuoteToASCII(s);
			} else {
				quoted = strconv.Quote(s);
			}
		}
		f.padString(quoted);
	};
	fmt.prototype.fmt_q = function(s) { return this.go$val.fmt_q(s); };
	fmt.Ptr.prototype.fmt_qc = function(c) {
		var f, quoted;
		f = this;
		quoted = (go$sliceType(Go$Uint8)).nil;
		if (f.plus) {
			quoted = strconv.AppendQuoteRuneToASCII(go$subslice(new (go$sliceType(Go$Uint8))(f.intbuf), 0, 0), ((c.low + ((c.high >> 31) * 4294967296)) >> 0));
		} else {
			quoted = strconv.AppendQuoteRune(go$subslice(new (go$sliceType(Go$Uint8))(f.intbuf), 0, 0), ((c.low + ((c.high >> 31) * 4294967296)) >> 0));
		}
		f.pad(quoted);
	};
	fmt.prototype.fmt_qc = function(c) { return this.go$val.fmt_qc(c); };
	doPrec = function(f, def) {
		if (f.precPresent) {
			return f.prec;
		}
		return def;
	};
	fmt.Ptr.prototype.formatFloat = function(v, verb, prec, n) {
		var f, slice, _ref;
		f = this;
		f.intbuf[0] = 32;
		slice = strconv.AppendFloat(go$subslice(new (go$sliceType(Go$Uint8))(f.intbuf), 0, 1), v, verb, prec, n);
		_ref = ((1 < 0 || 1 >= slice.length) ? go$throwRuntimeError("index out of range") : slice.array[slice.offset + 1]);
		if (_ref === 45 || _ref === 43) {
			if (f.zero && f.widPresent && f.wid > slice.length) {
				f.buf.WriteByte(((1 < 0 || 1 >= slice.length) ? go$throwRuntimeError("index out of range") : slice.array[slice.offset + 1]));
				f.wid = f.wid - 1 >> 0;
				f.pad(go$subslice(slice, 2));
				return;
			}
			slice = go$subslice(slice, 1);
		} else {
			if (f.plus) {
				(0 < 0 || 0 >= slice.length) ? go$throwRuntimeError("index out of range") : slice.array[slice.offset + 0] = 43;
			} else if (f.space) {
			} else {
				slice = go$subslice(slice, 1);
			}
		}
		f.pad(slice);
	};
	fmt.prototype.formatFloat = function(v, verb, prec, n) { return this.go$val.formatFloat(v, verb, prec, n); };
	fmt.Ptr.prototype.fmt_e64 = function(v) {
		var f;
		f = this;
		f.formatFloat(v, 101, doPrec(f, 6), 64);
	};
	fmt.prototype.fmt_e64 = function(v) { return this.go$val.fmt_e64(v); };
	fmt.Ptr.prototype.fmt_E64 = function(v) {
		var f;
		f = this;
		f.formatFloat(v, 69, doPrec(f, 6), 64);
	};
	fmt.prototype.fmt_E64 = function(v) { return this.go$val.fmt_E64(v); };
	fmt.Ptr.prototype.fmt_f64 = function(v) {
		var f;
		f = this;
		f.formatFloat(v, 102, doPrec(f, 6), 64);
	};
	fmt.prototype.fmt_f64 = function(v) { return this.go$val.fmt_f64(v); };
	fmt.Ptr.prototype.fmt_g64 = function(v) {
		var f;
		f = this;
		f.formatFloat(v, 103, doPrec(f, -1), 64);
	};
	fmt.prototype.fmt_g64 = function(v) { return this.go$val.fmt_g64(v); };
	fmt.Ptr.prototype.fmt_G64 = function(v) {
		var f;
		f = this;
		f.formatFloat(v, 71, doPrec(f, -1), 64);
	};
	fmt.prototype.fmt_G64 = function(v) { return this.go$val.fmt_G64(v); };
	fmt.Ptr.prototype.fmt_fb64 = function(v) {
		var f;
		f = this;
		f.formatFloat(v, 98, 0, 64);
	};
	fmt.prototype.fmt_fb64 = function(v) { return this.go$val.fmt_fb64(v); };
	fmt.Ptr.prototype.fmt_e32 = function(v) {
		var f;
		f = this;
		f.formatFloat(go$coerceFloat32(v), 101, doPrec(f, 6), 32);
	};
	fmt.prototype.fmt_e32 = function(v) { return this.go$val.fmt_e32(v); };
	fmt.Ptr.prototype.fmt_E32 = function(v) {
		var f;
		f = this;
		f.formatFloat(go$coerceFloat32(v), 69, doPrec(f, 6), 32);
	};
	fmt.prototype.fmt_E32 = function(v) { return this.go$val.fmt_E32(v); };
	fmt.Ptr.prototype.fmt_f32 = function(v) {
		var f;
		f = this;
		f.formatFloat(go$coerceFloat32(v), 102, doPrec(f, 6), 32);
	};
	fmt.prototype.fmt_f32 = function(v) { return this.go$val.fmt_f32(v); };
	fmt.Ptr.prototype.fmt_g32 = function(v) {
		var f;
		f = this;
		f.formatFloat(go$coerceFloat32(v), 103, doPrec(f, -1), 32);
	};
	fmt.prototype.fmt_g32 = function(v) { return this.go$val.fmt_g32(v); };
	fmt.Ptr.prototype.fmt_G32 = function(v) {
		var f;
		f = this;
		f.formatFloat(go$coerceFloat32(v), 71, doPrec(f, -1), 32);
	};
	fmt.prototype.fmt_G32 = function(v) { return this.go$val.fmt_G32(v); };
	fmt.Ptr.prototype.fmt_fb32 = function(v) {
		var f;
		f = this;
		f.formatFloat(go$coerceFloat32(v), 98, 0, 32);
	};
	fmt.prototype.fmt_fb32 = function(v) { return this.go$val.fmt_fb32(v); };
	fmt.Ptr.prototype.fmt_c64 = function(v, verb) {
		var f, r, oldPlus, i, _ref;
		f = this;
		f.buf.WriteByte(40);
		r = v.real;
		oldPlus = f.plus;
		i = 0;
		while (true) {
			_ref = verb;
			if (_ref === 98) {
				f.fmt_fb32(r);
			} else if (_ref === 101) {
				f.fmt_e32(r);
			} else if (_ref === 69) {
				f.fmt_E32(r);
			} else if (_ref === 102) {
				f.fmt_f32(r);
			} else if (_ref === 103) {
				f.fmt_g32(r);
			} else if (_ref === 71) {
				f.fmt_G32(r);
			}
			if (!((i === 0))) {
				break;
			}
			f.plus = true;
			r = v.imag;
			i = i + 1 >> 0;
		}
		f.plus = oldPlus;
		f.buf.Write(irparenBytes);
	};
	fmt.prototype.fmt_c64 = function(v, verb) { return this.go$val.fmt_c64(v, verb); };
	fmt.Ptr.prototype.fmt_c128 = function(v, verb) {
		var f, r, oldPlus, i, _ref;
		f = this;
		f.buf.WriteByte(40);
		r = v.real;
		oldPlus = f.plus;
		i = 0;
		while (true) {
			_ref = verb;
			if (_ref === 98) {
				f.fmt_fb64(r);
			} else if (_ref === 101) {
				f.fmt_e64(r);
			} else if (_ref === 69) {
				f.fmt_E64(r);
			} else if (_ref === 102) {
				f.fmt_f64(r);
			} else if (_ref === 103) {
				f.fmt_g64(r);
			} else if (_ref === 71) {
				f.fmt_G64(r);
			}
			if (!((i === 0))) {
				break;
			}
			f.plus = true;
			r = v.imag;
			i = i + 1 >> 0;
		}
		f.plus = oldPlus;
		f.buf.Write(irparenBytes);
	};
	fmt.prototype.fmt_c128 = function(v, verb) { return this.go$val.fmt_c128(v, verb); };
	go$ptrType(buffer).prototype.Write = function(p) {
		var n, err, b, _tmp, _tmp$1;
		n = 0;
		err = null;
		b = this;
		b.go$set(go$appendSlice(b.go$get(), p));
		_tmp = p.length; _tmp$1 = null; n = _tmp; err = _tmp$1;
		return [n, err];
	};
	buffer.prototype.Write = function(p) { var obj = this; return (new (go$ptrType(buffer))(function() { return obj; }, null)).Write(p); };
	go$ptrType(buffer).prototype.WriteString = function(s) {
		var n, err, b, _tmp, _tmp$1;
		n = 0;
		err = null;
		b = this;
		b.go$set(go$appendSlice(b.go$get(), new buffer(go$stringToBytes(s))));
		_tmp = s.length; _tmp$1 = null; n = _tmp; err = _tmp$1;
		return [n, err];
	};
	buffer.prototype.WriteString = function(s) { var obj = this; return (new (go$ptrType(buffer))(function() { return obj; }, null)).WriteString(s); };
	go$ptrType(buffer).prototype.WriteByte = function(c) {
		var b;
		b = this;
		b.go$set(go$append(b.go$get(), c));
		return null;
	};
	buffer.prototype.WriteByte = function(c) { var obj = this; return (new (go$ptrType(buffer))(function() { return obj; }, null)).WriteByte(c); };
	go$ptrType(buffer).prototype.WriteRune = function(r) {
		var bp, b, n, x, w;
		bp = this;
		if (r < 128) {
			bp.go$set(go$append(bp.go$get(), (r << 24 >>> 24)));
			return null;
		}
		b = bp.go$get();
		n = b.length;
		while ((n + 4 >> 0) > b.capacity) {
			b = go$append(b, 0);
		}
		w = utf8.EncodeRune((x = go$subslice(b, n, (n + 4 >> 0)), go$subslice(new (go$sliceType(Go$Uint8))(x.array), x.offset, x.offset + x.length)), r);
		bp.go$set(go$subslice(b, 0, (n + w >> 0)));
		return null;
	};
	buffer.prototype.WriteRune = function(r) { var obj = this; return (new (go$ptrType(buffer))(function() { return obj; }, null)).WriteRune(r); };
	cache.Ptr.prototype.put = function(x) {
		var c;
		c = this;
		c.mu.Lock();
		if (c.saved.length < c.saved.capacity) {
			c.saved = go$append(c.saved, x);
		}
		c.mu.Unlock();
	};
	cache.prototype.put = function(x) { return this.go$val.put(x); };
	cache.Ptr.prototype.get = function() {
		var c, n, x, x$1, x$2;
		c = this;
		c.mu.Lock();
		n = c.saved.length;
		if (n === 0) {
			c.mu.Unlock();
			return c.new$2();
		}
		x$2 = (x = c.saved, x$1 = n - 1 >> 0, ((x$1 < 0 || x$1 >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + x$1]));
		c.saved = go$subslice(c.saved, 0, (n - 1 >> 0));
		c.mu.Unlock();
		return x$2;
	};
	cache.prototype.get = function() { return this.go$val.get(); };
	newCache = function(f) {
		return new cache.Ptr(new sync.Mutex.Ptr(), (go$sliceType(go$emptyInterface)).make(0, 100, function() { return null; }), f);
	};
	newPrinter = function() {
		var x, p, v;
		p = (x = ppFree.get(), (x !== null && x.constructor === (go$ptrType(pp)) ? x.go$val : go$typeAssertionFailed(x, (go$ptrType(pp)))));
		p.panicking = false;
		p.erroring = false;
		p.fmt.init(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v;; }));
		return p;
	};
	pp.Ptr.prototype.free = function() {
		var p;
		p = this;
		if (p.buf.capacity > 1024) {
			return;
		}
		p.buf = go$subslice(p.buf, 0, 0);
		p.arg = null;
		p.value = new reflect.Value.Ptr((go$ptrType(reflect.rtype)).nil, 0, 0);
		ppFree.put(p);
	};
	pp.prototype.free = function() { return this.go$val.free(); };
	pp.Ptr.prototype.Width = function() {
		var wid, ok, p, _tmp, _tmp$1;
		wid = 0;
		ok = false;
		p = this;
		_tmp = p.fmt.wid; _tmp$1 = p.fmt.widPresent; wid = _tmp; ok = _tmp$1;
		return [wid, ok];
	};
	pp.prototype.Width = function() { return this.go$val.Width(); };
	pp.Ptr.prototype.Precision = function() {
		var prec, ok, p, _tmp, _tmp$1;
		prec = 0;
		ok = false;
		p = this;
		_tmp = p.fmt.prec; _tmp$1 = p.fmt.precPresent; prec = _tmp; ok = _tmp$1;
		return [prec, ok];
	};
	pp.prototype.Precision = function() { return this.go$val.Precision(); };
	pp.Ptr.prototype.Flag = function(b) {
		var p, _ref;
		p = this;
		_ref = b;
		if (_ref === 45) {
			return p.fmt.minus;
		} else if (_ref === 43) {
			return p.fmt.plus;
		} else if (_ref === 35) {
			return p.fmt.sharp;
		} else if (_ref === 32) {
			return p.fmt.space;
		} else if (_ref === 48) {
			return p.fmt.zero;
		}
		return false;
	};
	pp.prototype.Flag = function(b) { return this.go$val.Flag(b); };
	pp.Ptr.prototype.add = function(c) {
		var p, v;
		p = this;
		(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).WriteRune(c);
	};
	pp.prototype.add = function(c) { return this.go$val.add(c); };
	pp.Ptr.prototype.Write = function(b) {
		var ret, err, p, _tuple, v;
		ret = 0;
		err = null;
		p = this;
		_tuple = (new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).Write(b); ret = _tuple[0]; err = _tuple[1];
		return [ret, err];
	};
	pp.prototype.Write = function(b) { return this.go$val.Write(b); };
	Sprintf = go$pkg.Sprintf = function(format, a) {
		var p, s;
		p = newPrinter();
		p.doPrintf(format, a);
		s = go$bytesToString(p.buf);
		p.free();
		return s;
	};
	getField = function(v, i) {
		var _struct, val, _struct$1, _struct$2;
		val = (_struct = v.Field(i), new reflect.Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		if ((val.Kind() === 20) && !val.IsNil()) {
			val = (_struct$1 = val.Elem(), new reflect.Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
		}
		return (_struct$2 = val, new reflect.Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
	};
	parsenum = function(s, start, end) {
		var num, isnum, newi, _tmp, _tmp$1, _tmp$2;
		num = 0;
		isnum = false;
		newi = 0;
		if (start >= end) {
			_tmp = 0; _tmp$1 = false; _tmp$2 = end; num = _tmp; isnum = _tmp$1; newi = _tmp$2;
			return [num, isnum, newi];
		}
		newi = start;
		while (newi < end && 48 <= s.charCodeAt(newi) && s.charCodeAt(newi) <= 57) {
			num = ((((num >>> 16 << 16) * 10 >> 0) + (num << 16 >>> 16) * 10) >> 0) + ((s.charCodeAt(newi) - 48 << 24 >>> 24) >> 0) >> 0;
			isnum = true;
			newi = newi + 1 >> 0;
		}
		return [num, isnum, newi];
	};
	pp.Ptr.prototype.unknownType = function(v) {
		var p, v$1, v$2, v$3, v$4;
		p = this;
		if (go$interfaceIsEqual(v, null)) {
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).Write(nilAngleBytes);
			return;
		}
		(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).WriteByte(63);
		(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$3) { p.buf = v$3; })).WriteString(reflect.TypeOf(v).String());
		(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$4) { p.buf = v$4; })).WriteByte(63);
	};
	pp.prototype.unknownType = function(v) { return this.go$val.unknownType(v); };
	pp.Ptr.prototype.badVerb = function(verb) {
		var p, v, v$1, _struct, v$2;
		p = this;
		p.erroring = true;
		p.add(37);
		p.add(33);
		p.add(verb);
		p.add(40);
		if (!(go$interfaceIsEqual(p.arg, null))) {
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).WriteString(reflect.TypeOf(p.arg).String());
			p.add(61);
			p.printArg(p.arg, 118, false, false, 0);
		} else if (p.value.IsValid()) {
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).WriteString(p.value.Type().String());
			p.add(61);
			p.printValue((_struct = p.value, new reflect.Value.Ptr(_struct.typ, _struct.val, _struct.flag)), 118, false, false, 0);
		} else {
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).Write(nilAngleBytes);
		}
		p.add(41);
		p.erroring = false;
	};
	pp.prototype.badVerb = function(verb) { return this.go$val.badVerb(verb); };
	pp.Ptr.prototype.fmtBool = function(v, verb) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 116 || _ref === 118) {
			p.fmt.fmt_boolean(v);
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtBool = function(v, verb) { return this.go$val.fmtBool(v, verb); };
	pp.Ptr.prototype.fmtC = function(c) {
		var p, r, x, w;
		p = this;
		r = ((c.low + ((c.high >> 31) * 4294967296)) >> 0);
		if (!((x = new Go$Int64(0, r), (x.high === c.high && x.low === c.low)))) {
			r = 65533;
		}
		w = utf8.EncodeRune(go$subslice(new (go$sliceType(Go$Uint8))(p.runeBuf), 0, 4), r);
		p.fmt.pad(go$subslice(new (go$sliceType(Go$Uint8))(p.runeBuf), 0, w));
	};
	pp.prototype.fmtC = function(c) { return this.go$val.fmtC(c); };
	pp.Ptr.prototype.fmtInt64 = function(v, verb) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 98) {
			p.fmt.integer(v, new Go$Uint64(0, 2), true, "0123456789abcdef");
		} else if (_ref === 99) {
			p.fmtC(v);
		} else if (_ref === 100 || _ref === 118) {
			p.fmt.integer(v, new Go$Uint64(0, 10), true, "0123456789abcdef");
		} else if (_ref === 111) {
			p.fmt.integer(v, new Go$Uint64(0, 8), true, "0123456789abcdef");
		} else if (_ref === 113) {
			if ((0 < v.high || (0 === v.high && 0 <= v.low)) && (v.high < 0 || (v.high === 0 && v.low <= 1114111))) {
				p.fmt.fmt_qc(v);
			} else {
				p.badVerb(verb);
			}
		} else if (_ref === 120) {
			p.fmt.integer(v, new Go$Uint64(0, 16), true, "0123456789abcdef");
		} else if (_ref === 85) {
			p.fmtUnicode(v);
		} else if (_ref === 88) {
			p.fmt.integer(v, new Go$Uint64(0, 16), true, "0123456789ABCDEF");
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtInt64 = function(v, verb) { return this.go$val.fmtInt64(v, verb); };
	pp.Ptr.prototype.fmt0x64 = function(v, leading0x) {
		var p, sharp;
		p = this;
		sharp = p.fmt.sharp;
		p.fmt.sharp = leading0x;
		p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 16), false, "0123456789abcdef");
		p.fmt.sharp = sharp;
	};
	pp.prototype.fmt0x64 = function(v, leading0x) { return this.go$val.fmt0x64(v, leading0x); };
	pp.Ptr.prototype.fmtUnicode = function(v) {
		var p, precPresent, sharp, prec;
		p = this;
		precPresent = p.fmt.precPresent;
		sharp = p.fmt.sharp;
		p.fmt.sharp = false;
		prec = p.fmt.prec;
		if (!precPresent) {
			p.fmt.prec = 4;
			p.fmt.precPresent = true;
		}
		p.fmt.unicode = true;
		p.fmt.uniQuote = sharp;
		p.fmt.integer(v, new Go$Uint64(0, 16), false, "0123456789ABCDEF");
		p.fmt.unicode = false;
		p.fmt.uniQuote = false;
		p.fmt.prec = prec;
		p.fmt.precPresent = precPresent;
		p.fmt.sharp = sharp;
	};
	pp.prototype.fmtUnicode = function(v) { return this.go$val.fmtUnicode(v); };
	pp.Ptr.prototype.fmtUint64 = function(v, verb, goSyntax) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 98) {
			p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 2), false, "0123456789abcdef");
		} else if (_ref === 99) {
			p.fmtC(new Go$Int64(v.high, v.low));
		} else if (_ref === 100) {
			p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 10), false, "0123456789abcdef");
		} else if (_ref === 118) {
			if (goSyntax) {
				p.fmt0x64(v, true);
			} else {
				p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 10), false, "0123456789abcdef");
			}
		} else if (_ref === 111) {
			p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 8), false, "0123456789abcdef");
		} else if (_ref === 113) {
			if ((0 < v.high || (0 === v.high && 0 <= v.low)) && (v.high < 0 || (v.high === 0 && v.low <= 1114111))) {
				p.fmt.fmt_qc(new Go$Int64(v.high, v.low));
			} else {
				p.badVerb(verb);
			}
		} else if (_ref === 120) {
			p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 16), false, "0123456789abcdef");
		} else if (_ref === 88) {
			p.fmt.integer(new Go$Int64(v.high, v.low), new Go$Uint64(0, 16), false, "0123456789ABCDEF");
		} else if (_ref === 85) {
			p.fmtUnicode(new Go$Int64(v.high, v.low));
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtUint64 = function(v, verb, goSyntax) { return this.go$val.fmtUint64(v, verb, goSyntax); };
	pp.Ptr.prototype.fmtFloat32 = function(v, verb) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 98) {
			p.fmt.fmt_fb32(v);
		} else if (_ref === 101) {
			p.fmt.fmt_e32(v);
		} else if (_ref === 69) {
			p.fmt.fmt_E32(v);
		} else if (_ref === 102) {
			p.fmt.fmt_f32(v);
		} else if (_ref === 103 || _ref === 118) {
			p.fmt.fmt_g32(v);
		} else if (_ref === 71) {
			p.fmt.fmt_G32(v);
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtFloat32 = function(v, verb) { return this.go$val.fmtFloat32(v, verb); };
	pp.Ptr.prototype.fmtFloat64 = function(v, verb) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 98) {
			p.fmt.fmt_fb64(v);
		} else if (_ref === 101) {
			p.fmt.fmt_e64(v);
		} else if (_ref === 69) {
			p.fmt.fmt_E64(v);
		} else if (_ref === 102) {
			p.fmt.fmt_f64(v);
		} else if (_ref === 103 || _ref === 118) {
			p.fmt.fmt_g64(v);
		} else if (_ref === 71) {
			p.fmt.fmt_G64(v);
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtFloat64 = function(v, verb) { return this.go$val.fmtFloat64(v, verb); };
	pp.Ptr.prototype.fmtComplex64 = function(v, verb) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 98 || _ref === 101 || _ref === 69 || _ref === 102 || _ref === 70 || _ref === 103 || _ref === 71) {
			p.fmt.fmt_c64(v, verb);
		} else if (_ref === 118) {
			p.fmt.fmt_c64(v, 103);
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtComplex64 = function(v, verb) { return this.go$val.fmtComplex64(v, verb); };
	pp.Ptr.prototype.fmtComplex128 = function(v, verb) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 98 || _ref === 101 || _ref === 69 || _ref === 102 || _ref === 70 || _ref === 103 || _ref === 71) {
			p.fmt.fmt_c128(v, verb);
		} else if (_ref === 118) {
			p.fmt.fmt_c128(v, 103);
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtComplex128 = function(v, verb) { return this.go$val.fmtComplex128(v, verb); };
	pp.Ptr.prototype.fmtString = function(v, verb, goSyntax) {
		var p, _ref;
		p = this;
		_ref = verb;
		if (_ref === 118) {
			if (goSyntax) {
				p.fmt.fmt_q(v);
			} else {
				p.fmt.fmt_s(v);
			}
		} else if (_ref === 115) {
			p.fmt.fmt_s(v);
		} else if (_ref === 120) {
			p.fmt.fmt_sx(v, "0123456789abcdef");
		} else if (_ref === 88) {
			p.fmt.fmt_sx(v, "0123456789ABCDEF");
		} else if (_ref === 113) {
			p.fmt.fmt_q(v);
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtString = function(v, verb, goSyntax) { return this.go$val.fmtString(v, verb, goSyntax); };
	pp.Ptr.prototype.fmtBytes = function(v, verb, goSyntax, typ, depth) {
		var p, v$1, v$2, v$3, v$4, _ref, _i, c, i, v$5, v$6, v$7, v$8, _ref$1;
		p = this;
		if ((verb === 118) || (verb === 100)) {
			if (goSyntax) {
				if (go$interfaceIsEqual(typ, null)) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).Write(bytesBytes);
				} else {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).WriteString(typ.String());
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$3) { p.buf = v$3; })).WriteByte(123);
				}
			} else {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$4) { p.buf = v$4; })).WriteByte(91);
			}
			_ref = v;
			_i = 0;
			while (_i < _ref.length) {
				c = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
				i = _i;
				if (i > 0) {
					if (goSyntax) {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$5) { p.buf = v$5; })).Write(commaSpaceBytes);
					} else {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$6) { p.buf = v$6; })).WriteByte(32);
					}
				}
				p.printArg(new Go$Uint8(c), 118, p.fmt.plus, goSyntax, depth + 1 >> 0);
				_i++;
			}
			if (goSyntax) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$7) { p.buf = v$7; })).WriteByte(125);
			} else {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$8) { p.buf = v$8; })).WriteByte(93);
			}
			return;
		}
		_ref$1 = verb;
		if (_ref$1 === 115) {
			p.fmt.fmt_s(go$bytesToString(v));
		} else if (_ref$1 === 120) {
			p.fmt.fmt_bx(v, "0123456789abcdef");
		} else if (_ref$1 === 88) {
			p.fmt.fmt_bx(v, "0123456789ABCDEF");
		} else if (_ref$1 === 113) {
			p.fmt.fmt_q(go$bytesToString(v));
		} else {
			p.badVerb(verb);
		}
	};
	pp.prototype.fmtBytes = function(v, verb, goSyntax, typ, depth) { return this.go$val.fmtBytes(v, verb, goSyntax, typ, depth); };
	pp.Ptr.prototype.fmtPointer = function(value, verb, goSyntax) {
		var p, use0x64, _ref, u, _ref$1, v, v$1, v$2;
		p = this;
		use0x64 = true;
		_ref = verb;
		if (_ref === 112 || _ref === 118) {
		} else if (_ref === 98 || _ref === 100 || _ref === 111 || _ref === 120 || _ref === 88) {
			use0x64 = false;
		} else {
			p.badVerb(verb);
			return;
		}
		u = 0;
		_ref$1 = value.Kind();
		if (_ref$1 === 18 || _ref$1 === 19 || _ref$1 === 21 || _ref$1 === 22 || _ref$1 === 23 || _ref$1 === 26) {
			u = value.Pointer();
		} else {
			p.badVerb(verb);
			return;
		}
		if (goSyntax) {
			p.add(40);
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).WriteString(value.Type().String());
			p.add(41);
			p.add(40);
			if (u === 0) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).Write(nilBytes);
			} else {
				p.fmt0x64(new Go$Uint64(0, u.constructor === Number ? u : 1), true);
			}
			p.add(41);
		} else if ((verb === 118) && (u === 0)) {
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).Write(nilAngleBytes);
		} else {
			if (use0x64) {
				p.fmt0x64(new Go$Uint64(0, u.constructor === Number ? u : 1), !p.fmt.sharp);
			} else {
				p.fmtUint64(new Go$Uint64(0, u.constructor === Number ? u : 1), verb, false);
			}
		}
	};
	pp.prototype.fmtPointer = function(value, verb, goSyntax) { return this.go$val.fmtPointer(value, verb, goSyntax); };
	pp.Ptr.prototype.catchPanic = function(arg, verb) {
		var p, err, _struct, v, v$1, v$2, v$3, v$4;
		p = this;
		err = go$recover();
		if (!(go$interfaceIsEqual(err, null))) {
			v = (_struct = reflect.ValueOf(arg), new reflect.Value.Ptr(_struct.typ, _struct.val, _struct.flag));
			if ((v.Kind() === 22) && v.IsNil()) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).Write(nilAngleBytes);
				return;
			}
			if (p.panicking) {
				throw go$panic(err);
			}
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).Write(percentBangBytes);
			p.add(verb);
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$3) { p.buf = v$3; })).Write(panicBytes);
			p.panicking = true;
			p.printArg(err, 118, false, false, 0);
			p.panicking = false;
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$4) { p.buf = v$4; })).WriteByte(41);
		}
	};
	pp.prototype.catchPanic = function(arg, verb) { return this.go$val.catchPanic(arg, verb); };
	pp.Ptr.prototype.handleMethods = function(verb, plus, goSyntax, depth) {
		var wasString, handled, p, _tuple, x, formatter, ok, _tuple$1, x$1, stringer, ok$1, _ref, v, _ref$1, _type;
		wasString = false;
		handled = false;
		var go$deferred = [];
		try {
			p = this;
			if (p.erroring) {
				return [wasString, handled];
			}
			_tuple = (x = p.arg, (x !== null && Formatter.implementedBy.indexOf(x.constructor) !== -1 ? [x, true] : [null, false])); formatter = _tuple[0]; ok = _tuple[1];
			if (ok) {
				handled = true;
				wasString = false;
				go$deferred.push({ recv: p, method: "catchPanic", args: [p.arg, verb] });
				formatter.Format(p, verb);
				return [wasString, handled];
			}
			if (plus) {
				p.fmt.plus = false;
			}
			if (goSyntax) {
				p.fmt.sharp = false;
				_tuple$1 = (x$1 = p.arg, (x$1 !== null && GoStringer.implementedBy.indexOf(x$1.constructor) !== -1 ? [x$1, true] : [null, false])); stringer = _tuple$1[0]; ok$1 = _tuple$1[1];
				if (ok$1) {
					wasString = false;
					handled = true;
					go$deferred.push({ recv: p, method: "catchPanic", args: [p.arg, verb] });
					p.fmtString(stringer.GoString(), 115, false);
					return [wasString, handled];
				}
			} else {
				_ref = verb;
				if (_ref === 118 || _ref === 115 || _ref === 120 || _ref === 88 || _ref === 113) {
					_ref$1 = p.arg;
					_type = _ref$1 !== null ? _ref$1.constructor : null;
					if (go$error.implementedBy.indexOf(_type) !== -1) {
						v = _ref$1;
						wasString = false;
						handled = true;
						go$deferred.push({ recv: p, method: "catchPanic", args: [p.arg, verb] });
						p.printArg(new Go$String(v.Error()), verb, plus, false, depth);
						return [wasString, handled];
					} else if (Stringer.implementedBy.indexOf(_type) !== -1) {
						v = _ref$1;
						wasString = false;
						handled = true;
						go$deferred.push({ recv: p, method: "catchPanic", args: [p.arg, verb] });
						p.printArg(new Go$String(v.String()), verb, plus, false, depth);
						return [wasString, handled];
					}
				}
			}
			handled = false;
			return [wasString, handled];
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
			return [wasString, handled];
		}
	};
	pp.prototype.handleMethods = function(verb, plus, goSyntax, depth) { return this.go$val.handleMethods(verb, plus, goSyntax, depth); };
	pp.Ptr.prototype.printArg = function(arg, verb, plus, goSyntax, depth) {
		var wasString, p, _ref, _struct, oldPlus, oldSharp, f, _ref$1, _type, _tuple, isString, handled, _struct$1;
		wasString = false;
		p = this;
		p.arg = arg;
		p.value = new reflect.Value.Ptr((go$ptrType(reflect.rtype)).nil, 0, 0);
		if (go$interfaceIsEqual(arg, null)) {
			if ((verb === 84) || (verb === 118)) {
				p.fmt.pad(nilAngleBytes);
			} else {
				p.badVerb(verb);
			}
			wasString = false;
			return wasString;
		}
		_ref = verb;
		if (_ref === 84) {
			p.printArg(new Go$String(reflect.TypeOf(arg).String()), 115, false, false, 0);
			wasString = false;
			return wasString;
		} else if (_ref === 112) {
			p.fmtPointer((_struct = reflect.ValueOf(arg), new reflect.Value.Ptr(_struct.typ, _struct.val, _struct.flag)), verb, goSyntax);
			wasString = false;
			return wasString;
		}
		oldPlus = p.fmt.plus;
		oldSharp = p.fmt.sharp;
		if (plus) {
			p.fmt.plus = false;
		}
		if (goSyntax) {
			p.fmt.sharp = false;
		}
		_ref$1 = arg;
		_type = _ref$1 !== null ? _ref$1.constructor : null;
		if (_type === Go$Bool) {
			f = _ref$1.go$val;
			p.fmtBool(f, verb);
		} else if (_type === Go$Float32) {
			f = _ref$1.go$val;
			p.fmtFloat32(f, verb);
		} else if (_type === Go$Float64) {
			f = _ref$1.go$val;
			p.fmtFloat64(f, verb);
		} else if (_type === Go$Complex64) {
			f = _ref$1.go$val;
			p.fmtComplex64(f, verb);
		} else if (_type === Go$Complex128) {
			f = _ref$1.go$val;
			p.fmtComplex128(f, verb);
		} else if (_type === Go$Int) {
			f = _ref$1.go$val;
			p.fmtInt64(new Go$Int64(0, f), verb);
		} else if (_type === Go$Int8) {
			f = _ref$1.go$val;
			p.fmtInt64(new Go$Int64(0, f), verb);
		} else if (_type === Go$Int16) {
			f = _ref$1.go$val;
			p.fmtInt64(new Go$Int64(0, f), verb);
		} else if (_type === Go$Int32) {
			f = _ref$1.go$val;
			p.fmtInt64(new Go$Int64(0, f), verb);
		} else if (_type === Go$Int64) {
			f = _ref$1.go$val;
			p.fmtInt64(f, verb);
		} else if (_type === Go$Uint) {
			f = _ref$1.go$val;
			p.fmtUint64(new Go$Uint64(0, f), verb, goSyntax);
		} else if (_type === Go$Uint8) {
			f = _ref$1.go$val;
			p.fmtUint64(new Go$Uint64(0, f), verb, goSyntax);
		} else if (_type === Go$Uint16) {
			f = _ref$1.go$val;
			p.fmtUint64(new Go$Uint64(0, f), verb, goSyntax);
		} else if (_type === Go$Uint32) {
			f = _ref$1.go$val;
			p.fmtUint64(new Go$Uint64(0, f), verb, goSyntax);
		} else if (_type === Go$Uint64) {
			f = _ref$1.go$val;
			p.fmtUint64(f, verb, goSyntax);
		} else if (_type === Go$Uintptr) {
			f = _ref$1.go$val;
			p.fmtUint64(new Go$Uint64(0, f.constructor === Number ? f : 1), verb, goSyntax);
		} else if (_type === Go$String) {
			f = _ref$1.go$val;
			p.fmtString(f, verb, goSyntax);
			wasString = (verb === 115) || (verb === 118);
		} else if (_type === (go$sliceType(Go$Uint8))) {
			f = _ref$1.go$val;
			p.fmtBytes(f, verb, goSyntax, null, depth);
			wasString = verb === 115;
		} else {
			f = _ref$1;
			p.fmt.plus = oldPlus;
			p.fmt.sharp = oldSharp;
			_tuple = p.handleMethods(verb, plus, goSyntax, depth); isString = _tuple[0]; handled = _tuple[1];
			if (handled) {
				wasString = isString;
				return wasString;
			}
			wasString = p.printReflectValue((_struct$1 = reflect.ValueOf(arg), new reflect.Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), verb, plus, goSyntax, depth);
			return wasString;
		}
		p.arg = null;
		return wasString;
	};
	pp.prototype.printArg = function(arg, verb, plus, goSyntax, depth) { return this.go$val.printArg(arg, verb, plus, goSyntax, depth); };
	pp.Ptr.prototype.printValue = function(value, verb, plus, goSyntax, depth) {
		var wasString, p, v, _ref, _struct, _tuple, isString, handled, _struct$1;
		wasString = false;
		p = this;
		if (!value.IsValid()) {
			if ((verb === 84) || (verb === 118)) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).Write(nilAngleBytes);
			} else {
				p.badVerb(verb);
			}
			wasString = false;
			return wasString;
		}
		_ref = verb;
		if (_ref === 84) {
			p.printArg(new Go$String(value.Type().String()), 115, false, false, 0);
			wasString = false;
			return wasString;
		} else if (_ref === 112) {
			p.fmtPointer((_struct = value, new reflect.Value.Ptr(_struct.typ, _struct.val, _struct.flag)), verb, goSyntax);
			wasString = false;
			return wasString;
		}
		p.arg = null;
		if (value.CanInterface()) {
			p.arg = value.Interface();
		}
		_tuple = p.handleMethods(verb, plus, goSyntax, depth); isString = _tuple[0]; handled = _tuple[1];
		if (handled) {
			wasString = isString;
			return wasString;
		}
		wasString = p.printReflectValue((_struct$1 = value, new reflect.Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag)), verb, plus, goSyntax, depth);
		return wasString;
	};
	pp.prototype.printValue = function(value, verb, plus, goSyntax, depth) { return this.go$val.printValue(value, verb, plus, goSyntax, depth); };
	pp.Ptr.prototype.printReflectValue = function(value, verb, plus, goSyntax, depth) {
		var wasString, p, _struct, oldValue, _struct$1, _struct$2, f, _ref, x, v, v$1, v$2, v$3, keys, _ref$1, _i, _struct$3, key, i, v$4, v$5, _struct$4, v$6, _struct$5, _struct$6, v$7, v$8, v$9, _struct$7, v$10, t, i$1, v$11, v$12, _struct$8, f$1, v$13, v$14, _struct$9, _struct$10, v$15, _struct$11, value$1, v$16, v$17, v$18, _struct$12, typ, bytes, _ref$2, _i$1, i$2, v$19, v$20, v$21, v$22, i$3, v$23, v$24, _struct$13, v$25, v$26, v$27, _struct$14, a, _ref$3, v$28, _struct$15, v$29, _struct$16, _struct$17, _struct$18, _struct$19;
		wasString = false;
		p = this;
		oldValue = (_struct = p.value, new reflect.Value.Ptr(_struct.typ, _struct.val, _struct.flag));
		p.value = (_struct$1 = value, new reflect.Value.Ptr(_struct$1.typ, _struct$1.val, _struct$1.flag));
		f = (_struct$2 = value, new reflect.Value.Ptr(_struct$2.typ, _struct$2.val, _struct$2.flag));
		_ref = f.Kind();
		BigSwitch:
		switch (0) { default: if (_ref === 1) {
			p.fmtBool(f.Bool(), verb);
		} else if (_ref === 2 || _ref === 3 || _ref === 4 || _ref === 5 || _ref === 6) {
			p.fmtInt64(f.Int(), verb);
		} else if (_ref === 7 || _ref === 8 || _ref === 9 || _ref === 10 || _ref === 11 || _ref === 12) {
			p.fmtUint64(f.Uint(), verb, goSyntax);
		} else if (_ref === 13 || _ref === 14) {
			if (f.Type().Size() === 4) {
				p.fmtFloat32(f.Float(), verb);
			} else {
				p.fmtFloat64(f.Float(), verb);
			}
		} else if (_ref === 15 || _ref === 16) {
			if (f.Type().Size() === 8) {
				p.fmtComplex64((x = f.Complex(), new Go$Complex64(x.real, x.imag)), verb);
			} else {
				p.fmtComplex128(f.Complex(), verb);
			}
		} else if (_ref === 24) {
			p.fmtString(f.String(), verb, goSyntax);
		} else if (_ref === 21) {
			if (goSyntax) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).WriteString(f.Type().String());
				if (f.IsNil()) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).WriteString("(nil)");
					break;
				}
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).WriteByte(123);
			} else {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$3) { p.buf = v$3; })).Write(mapBytes);
			}
			keys = f.MapKeys();
			_ref$1 = keys;
			_i = 0;
			while (_i < _ref$1.length) {
				key = (_struct$3 = ((_i < 0 || _i >= _ref$1.length) ? go$throwRuntimeError("index out of range") : _ref$1.array[_ref$1.offset + _i]), new reflect.Value.Ptr(_struct$3.typ, _struct$3.val, _struct$3.flag));
				i = _i;
				if (i > 0) {
					if (goSyntax) {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$4) { p.buf = v$4; })).Write(commaSpaceBytes);
					} else {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$5) { p.buf = v$5; })).WriteByte(32);
					}
				}
				p.printValue((_struct$4 = key, new reflect.Value.Ptr(_struct$4.typ, _struct$4.val, _struct$4.flag)), verb, plus, goSyntax, depth + 1 >> 0);
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$6) { p.buf = v$6; })).WriteByte(58);
				p.printValue((_struct$6 = f.MapIndex((_struct$5 = key, new reflect.Value.Ptr(_struct$5.typ, _struct$5.val, _struct$5.flag))), new reflect.Value.Ptr(_struct$6.typ, _struct$6.val, _struct$6.flag)), verb, plus, goSyntax, depth + 1 >> 0);
				_i++;
			}
			if (goSyntax) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$7) { p.buf = v$7; })).WriteByte(125);
			} else {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$8) { p.buf = v$8; })).WriteByte(93);
			}
		} else if (_ref === 25) {
			if (goSyntax) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$9) { p.buf = v$9; })).WriteString(value.Type().String());
			}
			p.add(123);
			v$10 = (_struct$7 = f, new reflect.Value.Ptr(_struct$7.typ, _struct$7.val, _struct$7.flag));
			t = v$10.Type();
			i$1 = 0;
			while (i$1 < v$10.NumField()) {
				if (i$1 > 0) {
					if (goSyntax) {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$11) { p.buf = v$11; })).Write(commaSpaceBytes);
					} else {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$12) { p.buf = v$12; })).WriteByte(32);
					}
				}
				if (plus || goSyntax) {
					f$1 = (_struct$8 = t.Field(i$1), new reflect.StructField.Ptr(_struct$8.Name, _struct$8.PkgPath, _struct$8.Type, _struct$8.Tag, _struct$8.Offset, _struct$8.Index, _struct$8.Anonymous));
					if (!(f$1.Name === "")) {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$13) { p.buf = v$13; })).WriteString(f$1.Name);
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$14) { p.buf = v$14; })).WriteByte(58);
					}
				}
				p.printValue((_struct$10 = getField((_struct$9 = v$10, new reflect.Value.Ptr(_struct$9.typ, _struct$9.val, _struct$9.flag)), i$1), new reflect.Value.Ptr(_struct$10.typ, _struct$10.val, _struct$10.flag)), verb, plus, goSyntax, depth + 1 >> 0);
				i$1 = i$1 + 1 >> 0;
			}
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$15) { p.buf = v$15; })).WriteByte(125);
		} else if (_ref === 20) {
			value$1 = (_struct$11 = f.Elem(), new reflect.Value.Ptr(_struct$11.typ, _struct$11.val, _struct$11.flag));
			if (!value$1.IsValid()) {
				if (goSyntax) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$16) { p.buf = v$16; })).WriteString(f.Type().String());
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$17) { p.buf = v$17; })).Write(nilParenBytes);
				} else {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$18) { p.buf = v$18; })).Write(nilAngleBytes);
				}
			} else {
				wasString = p.printValue((_struct$12 = value$1, new reflect.Value.Ptr(_struct$12.typ, _struct$12.val, _struct$12.flag)), verb, plus, goSyntax, depth + 1 >> 0);
			}
		} else if (_ref === 17 || _ref === 23) {
			typ = f.Type();
			if (typ.Elem().Kind() === 8) {
				bytes = (go$sliceType(Go$Uint8)).nil;
				if (f.Kind() === 23) {
					bytes = f.Bytes();
				} else if (f.CanAddr()) {
					bytes = f.Slice(0, f.Len()).Bytes();
				} else {
					bytes = (go$sliceType(Go$Uint8)).make(f.Len(), 0, function() { return 0; });
					_ref$2 = bytes;
					_i$1 = 0;
					while (_i$1 < _ref$2.length) {
						i$2 = _i$1;
						(i$2 < 0 || i$2 >= bytes.length) ? go$throwRuntimeError("index out of range") : bytes.array[bytes.offset + i$2] = (f.Index(i$2).Uint().low << 24 >>> 24);
						_i$1++;
					}
				}
				p.fmtBytes(bytes, verb, goSyntax, typ, depth);
				wasString = verb === 115;
				break;
			}
			if (goSyntax) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$19) { p.buf = v$19; })).WriteString(value.Type().String());
				if ((f.Kind() === 23) && f.IsNil()) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$20) { p.buf = v$20; })).WriteString("(nil)");
					break;
				}
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$21) { p.buf = v$21; })).WriteByte(123);
			} else {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$22) { p.buf = v$22; })).WriteByte(91);
			}
			i$3 = 0;
			while (i$3 < f.Len()) {
				if (i$3 > 0) {
					if (goSyntax) {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$23) { p.buf = v$23; })).Write(commaSpaceBytes);
					} else {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$24) { p.buf = v$24; })).WriteByte(32);
					}
				}
				p.printValue((_struct$13 = f.Index(i$3), new reflect.Value.Ptr(_struct$13.typ, _struct$13.val, _struct$13.flag)), verb, plus, goSyntax, depth + 1 >> 0);
				i$3 = i$3 + 1 >> 0;
			}
			if (goSyntax) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$25) { p.buf = v$25; })).WriteByte(125);
			} else {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$26) { p.buf = v$26; })).WriteByte(93);
			}
		} else if (_ref === 22) {
			v$27 = f.Pointer();
			if (!((v$27 === 0)) && (depth === 0)) {
				a = (_struct$14 = f.Elem(), new reflect.Value.Ptr(_struct$14.typ, _struct$14.val, _struct$14.flag));
				_ref$3 = a.Kind();
				if (_ref$3 === 17 || _ref$3 === 23) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$28) { p.buf = v$28; })).WriteByte(38);
					p.printValue((_struct$15 = a, new reflect.Value.Ptr(_struct$15.typ, _struct$15.val, _struct$15.flag)), verb, plus, goSyntax, depth + 1 >> 0);
					break BigSwitch;
				} else if (_ref$3 === 25) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$29) { p.buf = v$29; })).WriteByte(38);
					p.printValue((_struct$16 = a, new reflect.Value.Ptr(_struct$16.typ, _struct$16.val, _struct$16.flag)), verb, plus, goSyntax, depth + 1 >> 0);
					break BigSwitch;
				}
			}
			p.fmtPointer((_struct$17 = value, new reflect.Value.Ptr(_struct$17.typ, _struct$17.val, _struct$17.flag)), verb, goSyntax);
		} else if (_ref === 18 || _ref === 19 || _ref === 26) {
			p.fmtPointer((_struct$18 = value, new reflect.Value.Ptr(_struct$18.typ, _struct$18.val, _struct$18.flag)), verb, goSyntax);
		} else {
			p.unknownType(new f.constructor.Struct(f));
		} }
		p.value = (_struct$19 = oldValue, new reflect.Value.Ptr(_struct$19.typ, _struct$19.val, _struct$19.flag));
		wasString = wasString;
		return wasString;
	};
	pp.prototype.printReflectValue = function(value, verb, plus, goSyntax, depth) { return this.go$val.printReflectValue(value, verb, plus, goSyntax, depth); };
	intFromArg = function(a, argNum) {
		var num, isInt, newArgNum, _tuple, x;
		num = 0;
		isInt = false;
		newArgNum = 0;
		newArgNum = argNum;
		if (argNum < a.length) {
			_tuple = (x = ((argNum < 0 || argNum >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + argNum]), (x !== null && x.constructor === Go$Int ? [x.go$val, true] : [0, false])); num = _tuple[0]; isInt = _tuple[1];
			newArgNum = argNum + 1 >> 0;
		}
		return [num, isInt, newArgNum];
	};
	parseArgNumber = function(format) {
		var index, wid, ok, i, _tuple, width, ok$1, newi, _tmp, _tmp$1, _tmp$2, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, _tmp$8;
		index = 0;
		wid = 0;
		ok = false;
		i = 1;
		while (i < format.length) {
			if (format.charCodeAt(i) === 93) {
				_tuple = parsenum(format, 1, i); width = _tuple[0]; ok$1 = _tuple[1]; newi = _tuple[2];
				if (!ok$1 || !((newi === i))) {
					_tmp = 0; _tmp$1 = i + 1 >> 0; _tmp$2 = false; index = _tmp; wid = _tmp$1; ok = _tmp$2;
					return [index, wid, ok];
				}
				_tmp$3 = width - 1 >> 0; _tmp$4 = i + 1 >> 0; _tmp$5 = true; index = _tmp$3; wid = _tmp$4; ok = _tmp$5;
				return [index, wid, ok];
			}
			i = i + 1 >> 0;
		}
		_tmp$6 = 0; _tmp$7 = 1; _tmp$8 = false; index = _tmp$6; wid = _tmp$7; ok = _tmp$8;
		return [index, wid, ok];
	};
	pp.Ptr.prototype.argNumber = function(argNum, format, i, numArgs) {
		var newArgNum, newi, found, p, _tmp, _tmp$1, _tmp$2, _tuple, index, wid, ok, _tmp$3, _tmp$4, _tmp$5, _tmp$6, _tmp$7, _tmp$8;
		newArgNum = 0;
		newi = 0;
		found = false;
		p = this;
		if (format.length <= i || !((format.charCodeAt(i) === 91))) {
			_tmp = argNum; _tmp$1 = i; _tmp$2 = false; newArgNum = _tmp; newi = _tmp$1; found = _tmp$2;
			return [newArgNum, newi, found];
		}
		p.reordered = true;
		_tuple = parseArgNumber(format.substring(i)); index = _tuple[0]; wid = _tuple[1]; ok = _tuple[2];
		if (ok && 0 <= index && index < numArgs) {
			_tmp$3 = index; _tmp$4 = i + wid >> 0; _tmp$5 = true; newArgNum = _tmp$3; newi = _tmp$4; found = _tmp$5;
			return [newArgNum, newi, found];
		}
		p.goodArgNum = false;
		_tmp$6 = argNum; _tmp$7 = i + wid >> 0; _tmp$8 = true; newArgNum = _tmp$6; newi = _tmp$7; found = _tmp$8;
		return [newArgNum, newi, found];
	};
	pp.prototype.argNumber = function(argNum, format, i, numArgs) { return this.go$val.argNumber(argNum, format, i, numArgs); };
	pp.Ptr.prototype.doPrintf = function(format, a) {
		var p, end, argNum, afterIndex, i, lasti, v, _ref, _tuple, _tuple$1, v$1, _tuple$2, _tuple$3, _tuple$4, v$2, _tuple$5, _tuple$6, v$3, _tuple$7, c, w, v$4, v$5, v$6, v$7, v$8, arg, goSyntax, plus, v$9, arg$1, v$10, v$11, v$12, v$13;
		p = this;
		end = format.length;
		argNum = 0;
		afterIndex = false;
		p.reordered = false;
		i = 0;
		while (i < end) {
			p.goodArgNum = true;
			lasti = i;
			while (i < end && !((format.charCodeAt(i) === 37))) {
				i = i + 1 >> 0;
			}
			if (i > lasti) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v) { p.buf = v; })).WriteString(format.substring(lasti, i));
			}
			if (i >= end) {
				break;
			}
			i = i + 1 >> 0;
			p.fmt.clearflags();
			F:
			while (i < end) {
				_ref = format.charCodeAt(i);
				if (_ref === 35) {
					p.fmt.sharp = true;
				} else if (_ref === 48) {
					p.fmt.zero = true;
				} else if (_ref === 43) {
					p.fmt.plus = true;
				} else if (_ref === 45) {
					p.fmt.minus = true;
				} else if (_ref === 32) {
					p.fmt.space = true;
				} else {
					break F;
				}
				i = i + 1 >> 0;
			}
			_tuple = p.argNumber(argNum, format, i, a.length); argNum = _tuple[0]; i = _tuple[1]; afterIndex = _tuple[2];
			if (i < end && (format.charCodeAt(i) === 42)) {
				i = i + 1 >> 0;
				_tuple$1 = intFromArg(a, argNum); p.fmt.wid = _tuple$1[0]; p.fmt.widPresent = _tuple$1[1]; argNum = _tuple$1[2];
				if (!p.fmt.widPresent) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$1) { p.buf = v$1; })).Write(badWidthBytes);
				}
				afterIndex = false;
			} else {
				_tuple$2 = parsenum(format, i, end); p.fmt.wid = _tuple$2[0]; p.fmt.widPresent = _tuple$2[1]; i = _tuple$2[2];
				if (afterIndex && p.fmt.widPresent) {
					p.goodArgNum = false;
				}
			}
			if ((i + 1 >> 0) < end && (format.charCodeAt(i) === 46)) {
				i = i + 1 >> 0;
				if (afterIndex) {
					p.goodArgNum = false;
				}
				_tuple$3 = p.argNumber(argNum, format, i, a.length); argNum = _tuple$3[0]; i = _tuple$3[1]; afterIndex = _tuple$3[2];
				if (format.charCodeAt(i) === 42) {
					i = i + 1 >> 0;
					_tuple$4 = intFromArg(a, argNum); p.fmt.prec = _tuple$4[0]; p.fmt.precPresent = _tuple$4[1]; argNum = _tuple$4[2];
					if (!p.fmt.precPresent) {
						(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$2) { p.buf = v$2; })).Write(badPrecBytes);
					}
					afterIndex = false;
				} else {
					_tuple$5 = parsenum(format, i, end); p.fmt.prec = _tuple$5[0]; p.fmt.precPresent = _tuple$5[1]; i = _tuple$5[2];
					if (!p.fmt.precPresent) {
						p.fmt.prec = 0;
						p.fmt.precPresent = true;
					}
				}
			}
			if (!afterIndex) {
				_tuple$6 = p.argNumber(argNum, format, i, a.length); argNum = _tuple$6[0]; i = _tuple$6[1]; afterIndex = _tuple$6[2];
			}
			if (i >= end) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$3) { p.buf = v$3; })).Write(noVerbBytes);
				continue;
			}
			_tuple$7 = utf8.DecodeRuneInString(format.substring(i)); c = _tuple$7[0]; w = _tuple$7[1];
			i = i + (w) >> 0;
			if (c === 37) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$4) { p.buf = v$4; })).WriteByte(37);
				continue;
			}
			if (!p.goodArgNum) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$5) { p.buf = v$5; })).Write(percentBangBytes);
				p.add(c);
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$6) { p.buf = v$6; })).Write(badIndexBytes);
				continue;
			} else if (argNum >= a.length) {
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$7) { p.buf = v$7; })).Write(percentBangBytes);
				p.add(c);
				(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$8) { p.buf = v$8; })).Write(missingBytes);
				continue;
			}
			arg = ((argNum < 0 || argNum >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + argNum]);
			argNum = argNum + 1 >> 0;
			goSyntax = (c === 118) && p.fmt.sharp;
			plus = (c === 118) && p.fmt.plus;
			p.printArg(arg, c, plus, goSyntax, 0);
		}
		if (!p.reordered && argNum < a.length) {
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$9) { p.buf = v$9; })).Write(extraBytes);
			while (argNum < a.length) {
				arg$1 = ((argNum < 0 || argNum >= a.length) ? go$throwRuntimeError("index out of range") : a.array[a.offset + argNum]);
				if (!(go$interfaceIsEqual(arg$1, null))) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$10) { p.buf = v$10; })).WriteString(reflect.TypeOf(arg$1).String());
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$11) { p.buf = v$11; })).WriteByte(61);
				}
				p.printArg(arg$1, 118, false, false, 0);
				if ((argNum + 1 >> 0) < a.length) {
					(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$12) { p.buf = v$12; })).Write(commaSpaceBytes);
				}
				argNum = argNum + 1 >> 0;
			}
			(new (go$ptrType(buffer))(function() { return p.buf; }, function(v$13) { p.buf = v$13; })).WriteByte(41);
		}
	};
	pp.prototype.doPrintf = function(format, a) { return this.go$val.doPrintf(format, a); };
	ss.Ptr.prototype.Read = function(buf) {
		var n, err, s, _tmp, _tmp$1;
		n = 0;
		err = null;
		s = this;
		_tmp = 0; _tmp$1 = errors.New("ScanState's Read should not be called. Use ReadRune"); n = _tmp; err = _tmp$1;
		return [n, err];
	};
	ss.prototype.Read = function(buf) { return this.go$val.Read(buf); };
	ss.Ptr.prototype.ReadRune = function() {
		var r, size, err, s, _tuple;
		r = 0;
		size = 0;
		err = null;
		s = this;
		if (s.peekRune >= 0) {
			s.count = s.count + 1 >> 0;
			r = s.peekRune;
			size = utf8.RuneLen(r);
			s.prevRune = r;
			s.peekRune = -1;
			return [r, size, err];
		}
		if (s.atEOF || s.ssave.nlIsEnd && (s.prevRune === 10) || s.count >= s.ssave.argLimit) {
			err = io.EOF;
			return [r, size, err];
		}
		_tuple = s.rr.ReadRune(); r = _tuple[0]; size = _tuple[1]; err = _tuple[2];
		if (go$interfaceIsEqual(err, null)) {
			s.count = s.count + 1 >> 0;
			s.prevRune = r;
		} else if (go$interfaceIsEqual(err, io.EOF)) {
			s.atEOF = true;
		}
		return [r, size, err];
	};
	ss.prototype.ReadRune = function() { return this.go$val.ReadRune(); };
	ss.Ptr.prototype.Width = function() {
		var wid, ok, s, _tmp, _tmp$1, _tmp$2, _tmp$3;
		wid = 0;
		ok = false;
		s = this;
		if (s.ssave.maxWid === 1073741824) {
			_tmp = 0; _tmp$1 = false; wid = _tmp; ok = _tmp$1;
			return [wid, ok];
		}
		_tmp$2 = s.ssave.maxWid; _tmp$3 = true; wid = _tmp$2; ok = _tmp$3;
		return [wid, ok];
	};
	ss.prototype.Width = function() { return this.go$val.Width(); };
	ss.Ptr.prototype.getRune = function() {
		var r, s, _tuple, err;
		r = 0;
		s = this;
		_tuple = s.ReadRune(); r = _tuple[0]; err = _tuple[2];
		if (!(go$interfaceIsEqual(err, null))) {
			if (go$interfaceIsEqual(err, io.EOF)) {
				r = -1;
				return r;
			}
			s.error(err);
		}
		return r;
	};
	ss.prototype.getRune = function() { return this.go$val.getRune(); };
	ss.Ptr.prototype.UnreadRune = function() {
		var s, _tuple, x, u, ok;
		s = this;
		_tuple = (x = s.rr, (x !== null && runeUnreader.implementedBy.indexOf(x.constructor) !== -1 ? [x, true] : [null, false])); u = _tuple[0]; ok = _tuple[1];
		if (ok) {
			u.UnreadRune();
		} else {
			s.peekRune = s.prevRune;
		}
		s.prevRune = -1;
		s.count = s.count - 1 >> 0;
		return null;
	};
	ss.prototype.UnreadRune = function() { return this.go$val.UnreadRune(); };
	ss.Ptr.prototype.error = function(err) {
		var s, x;
		s = this;
		throw go$panic((x = new scanError.Ptr(err), new x.constructor.Struct(x)));
	};
	ss.prototype.error = function(err) { return this.go$val.error(err); };
	ss.Ptr.prototype.errorString = function(err) {
		var s, x;
		s = this;
		throw go$panic((x = new scanError.Ptr(errors.New(err)), new x.constructor.Struct(x)));
	};
	ss.prototype.errorString = function(err) { return this.go$val.errorString(err); };
	ss.Ptr.prototype.Token = function(skipSpace, f) {
		var tok, err, s;
		tok = (go$sliceType(Go$Uint8)).nil;
		err = null;
		var go$deferred = [];
		try {
			s = this;
			go$deferred.push({ fun: (function() {
				var e, _tuple, _struct, se, ok;
				e = go$recover();
				if (!(go$interfaceIsEqual(e, null))) {
					_tuple = (e !== null && e.constructor === scanError ? [e.go$val, true] : [new scanError.Ptr(), false]); se = (_struct = _tuple[0], new scanError.Ptr(_struct.err)); ok = _tuple[1];
					if (ok) {
						err = se.err;
					} else {
						throw go$panic(e);
					}
				}
			}), args: [] });
			if (f === go$throwNilPointerError) {
				f = notSpace;
			}
			s.buf = go$subslice(s.buf, 0, 0);
			tok = s.token(skipSpace, f);
			return [tok, err];
		} catch(go$err) {
			go$pushErr(go$err);
		} finally {
			go$callDeferred(go$deferred);
			return [tok, err];
		}
	};
	ss.prototype.Token = function(skipSpace, f) { return this.go$val.Token(skipSpace, f); };
	isSpace = function(r) {
		var rx, _ref, _i, rng;
		if (r >= 65536) {
			return false;
		}
		rx = (r << 16 >>> 16);
		_ref = space;
		_i = 0;
		while (_i < _ref.length) {
			rng = go$mapArray(((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]), function(entry) { return entry; });
			if (rx < rng[0]) {
				return false;
			}
			if (rx <= rng[1]) {
				return true;
			}
			_i++;
		}
		return false;
	};
	notSpace = function(r) {
		return !isSpace(r);
	};
	ss.Ptr.prototype.SkipSpace = function() {
		var s;
		s = this;
		s.skipSpace(false);
	};
	ss.prototype.SkipSpace = function() { return this.go$val.SkipSpace(); };
	ss.Ptr.prototype.free = function(old) {
		var s, _struct;
		s = this;
		if (old.validSave) {
			s.ssave = (_struct = old, new ssave.Ptr(_struct.validSave, _struct.nlIsEnd, _struct.nlIsSpace, _struct.argLimit, _struct.limit, _struct.maxWid));
			return;
		}
		if (s.buf.capacity > 1024) {
			return;
		}
		s.buf = go$subslice(s.buf, 0, 0);
		s.rr = null;
		ssFree.put(s);
	};
	ss.prototype.free = function(old) { return this.go$val.free(old); };
	ss.Ptr.prototype.skipSpace = function(stopAtNewline) {
		var s, r;
		s = this;
		while (true) {
			r = s.getRune();
			if (r === -1) {
				return;
			}
			if ((r === 13) && s.peek("\n")) {
				continue;
			}
			if (r === 10) {
				if (stopAtNewline) {
					break;
				}
				if (s.ssave.nlIsSpace) {
					continue;
				}
				s.errorString("unexpected newline");
				return;
			}
			if (!isSpace(r)) {
				s.UnreadRune();
				break;
			}
		}
	};
	ss.prototype.skipSpace = function(stopAtNewline) { return this.go$val.skipSpace(stopAtNewline); };
	ss.Ptr.prototype.token = function(skipSpace, f) {
		var s, r, v, x;
		s = this;
		if (skipSpace) {
			s.skipSpace(false);
		}
		while (true) {
			r = s.getRune();
			if (r === -1) {
				break;
			}
			if (!f(r)) {
				s.UnreadRune();
				break;
			}
			(new (go$ptrType(buffer))(function() { return s.buf; }, function(v) { s.buf = v; })).WriteRune(r);
		}
		return (x = s.buf, go$subslice(new (go$sliceType(Go$Uint8))(x.array), x.offset, x.offset + x.length));
	};
	ss.prototype.token = function(skipSpace, f) { return this.go$val.token(skipSpace, f); };
	indexRune = function(s, r) {
		var _ref, _i, _rune, c, i;
		_ref = s;
		_i = 0;
		while (_i < _ref.length) {
			_rune = go$decodeRune(_ref, _i);
			c = _rune[0];
			i = _i;
			if (c === r) {
				return i;
			}
			_i += _rune[1];
		}
		return -1;
	};
	ss.Ptr.prototype.peek = function(ok) {
		var s, r;
		s = this;
		r = s.getRune();
		if (!((r === -1))) {
			s.UnreadRune();
		}
		return indexRune(ok, r) >= 0;
	};
	ss.prototype.peek = function(ok) { return this.go$val.peek(ok); };
	go$pkg.init = function() {
		(go$ptrType(fmt)).methods = [["clearflags", "fmt", [], [], false, -1], ["computePadding", "fmt", [Go$Int], [(go$sliceType(Go$Uint8)), Go$Int, Go$Int], false, -1], ["fmt_E32", "fmt", [Go$Float32], [], false, -1], ["fmt_E64", "fmt", [Go$Float64], [], false, -1], ["fmt_G32", "fmt", [Go$Float32], [], false, -1], ["fmt_G64", "fmt", [Go$Float64], [], false, -1], ["fmt_boolean", "fmt", [Go$Bool], [], false, -1], ["fmt_bx", "fmt", [(go$sliceType(Go$Uint8)), Go$String], [], false, -1], ["fmt_c128", "fmt", [Go$Complex128, Go$Int32], [], false, -1], ["fmt_c64", "fmt", [Go$Complex64, Go$Int32], [], false, -1], ["fmt_e32", "fmt", [Go$Float32], [], false, -1], ["fmt_e64", "fmt", [Go$Float64], [], false, -1], ["fmt_f32", "fmt", [Go$Float32], [], false, -1], ["fmt_f64", "fmt", [Go$Float64], [], false, -1], ["fmt_fb32", "fmt", [Go$Float32], [], false, -1], ["fmt_fb64", "fmt", [Go$Float64], [], false, -1], ["fmt_g32", "fmt", [Go$Float32], [], false, -1], ["fmt_g64", "fmt", [Go$Float64], [], false, -1], ["fmt_q", "fmt", [Go$String], [], false, -1], ["fmt_qc", "fmt", [Go$Int64], [], false, -1], ["fmt_s", "fmt", [Go$String], [], false, -1], ["fmt_sbx", "fmt", [Go$String, (go$sliceType(Go$Uint8)), Go$String], [], false, -1], ["fmt_sx", "fmt", [Go$String, Go$String], [], false, -1], ["formatFloat", "fmt", [Go$Float64, Go$Uint8, Go$Int, Go$Int], [], false, -1], ["init", "fmt", [(go$ptrType(buffer))], [], false, -1], ["integer", "fmt", [Go$Int64, Go$Uint64, Go$Bool, Go$String], [], false, -1], ["pad", "fmt", [(go$sliceType(Go$Uint8))], [], false, -1], ["padString", "fmt", [Go$String], [], false, -1], ["truncate", "fmt", [Go$String], [Go$String], false, -1], ["writePadding", "fmt", [Go$Int, (go$sliceType(Go$Uint8))], [], false, -1]];
		fmt.init([["intbuf", "intbuf", "fmt", (go$arrayType(Go$Uint8, 65)), ""], ["buf", "buf", "fmt", (go$ptrType(buffer)), ""], ["wid", "wid", "fmt", Go$Int, ""], ["prec", "prec", "fmt", Go$Int, ""], ["widPresent", "widPresent", "fmt", Go$Bool, ""], ["precPresent", "precPresent", "fmt", Go$Bool, ""], ["minus", "minus", "fmt", Go$Bool, ""], ["plus", "plus", "fmt", Go$Bool, ""], ["sharp", "sharp", "fmt", Go$Bool, ""], ["space", "space", "fmt", Go$Bool, ""], ["unicode", "unicode", "fmt", Go$Bool, ""], ["uniQuote", "uniQuote", "fmt", Go$Bool, ""], ["zero", "zero", "fmt", Go$Bool, ""]]);
		State.init([["Flag", "", (go$funcType([Go$Int], [Go$Bool], false))], ["Precision", "", (go$funcType([], [Go$Int, Go$Bool], false))], ["Width", "", (go$funcType([], [Go$Int, Go$Bool], false))], ["Write", "", (go$funcType([(go$sliceType(Go$Uint8))], [Go$Int, go$error], false))]]);
		Formatter.init([["Format", "", (go$funcType([State, Go$Int32], [], false))]]);
		Stringer.init([["String", "", (go$funcType([], [Go$String], false))]]);
		GoStringer.init([["GoString", "", (go$funcType([], [Go$String], false))]]);
		(go$ptrType(buffer)).methods = [["Write", "", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1], ["WriteByte", "", [Go$Uint8], [go$error], false, -1], ["WriteRune", "", [Go$Int32], [go$error], false, -1], ["WriteString", "", [Go$String], [Go$Int, go$error], false, -1]];
		buffer.init(Go$Uint8);
		(go$ptrType(pp)).methods = [["Flag", "", [Go$Int], [Go$Bool], false, -1], ["Precision", "", [], [Go$Int, Go$Bool], false, -1], ["Width", "", [], [Go$Int, Go$Bool], false, -1], ["Write", "", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1], ["add", "fmt", [Go$Int32], [], false, -1], ["argNumber", "fmt", [Go$Int, Go$String, Go$Int, Go$Int], [Go$Int, Go$Int, Go$Bool], false, -1], ["badVerb", "fmt", [Go$Int32], [], false, -1], ["catchPanic", "fmt", [go$emptyInterface, Go$Int32], [], false, -1], ["doPrint", "fmt", [(go$sliceType(go$emptyInterface)), Go$Bool, Go$Bool], [], false, -1], ["doPrintf", "fmt", [Go$String, (go$sliceType(go$emptyInterface))], [], false, -1], ["fmt0x64", "fmt", [Go$Uint64, Go$Bool], [], false, -1], ["fmtBool", "fmt", [Go$Bool, Go$Int32], [], false, -1], ["fmtBytes", "fmt", [(go$sliceType(Go$Uint8)), Go$Int32, Go$Bool, reflect.Type, Go$Int], [], false, -1], ["fmtC", "fmt", [Go$Int64], [], false, -1], ["fmtComplex128", "fmt", [Go$Complex128, Go$Int32], [], false, -1], ["fmtComplex64", "fmt", [Go$Complex64, Go$Int32], [], false, -1], ["fmtFloat32", "fmt", [Go$Float32, Go$Int32], [], false, -1], ["fmtFloat64", "fmt", [Go$Float64, Go$Int32], [], false, -1], ["fmtInt64", "fmt", [Go$Int64, Go$Int32], [], false, -1], ["fmtPointer", "fmt", [reflect.Value, Go$Int32, Go$Bool], [], false, -1], ["fmtString", "fmt", [Go$String, Go$Int32, Go$Bool], [], false, -1], ["fmtUint64", "fmt", [Go$Uint64, Go$Int32, Go$Bool], [], false, -1], ["fmtUnicode", "fmt", [Go$Int64], [], false, -1], ["free", "fmt", [], [], false, -1], ["handleMethods", "fmt", [Go$Int32, Go$Bool, Go$Bool, Go$Int], [Go$Bool, Go$Bool], false, -1], ["printArg", "fmt", [go$emptyInterface, Go$Int32, Go$Bool, Go$Bool, Go$Int], [Go$Bool], false, -1], ["printReflectValue", "fmt", [reflect.Value, Go$Int32, Go$Bool, Go$Bool, Go$Int], [Go$Bool], false, -1], ["printValue", "fmt", [reflect.Value, Go$Int32, Go$Bool, Go$Bool, Go$Int], [Go$Bool], false, -1], ["unknownType", "fmt", [go$emptyInterface], [], false, -1]];
		pp.init([["n", "n", "fmt", Go$Int, ""], ["panicking", "panicking", "fmt", Go$Bool, ""], ["erroring", "erroring", "fmt", Go$Bool, ""], ["buf", "buf", "fmt", buffer, ""], ["arg", "arg", "fmt", go$emptyInterface, ""], ["value", "value", "fmt", reflect.Value, ""], ["reordered", "reordered", "fmt", Go$Bool, ""], ["goodArgNum", "goodArgNum", "fmt", Go$Bool, ""], ["runeBuf", "runeBuf", "fmt", (go$arrayType(Go$Uint8, 4)), ""], ["fmt", "fmt", "fmt", fmt, ""]]);
		(go$ptrType(cache)).methods = [["get", "fmt", [], [go$emptyInterface], false, -1], ["put", "fmt", [go$emptyInterface], [], false, -1]];
		cache.init([["mu", "mu", "fmt", sync.Mutex, ""], ["saved", "saved", "fmt", (go$sliceType(go$emptyInterface)), ""], ["new$2", "new", "fmt", (go$funcType([], [go$emptyInterface], false)), ""]]);
		runeUnreader.init([["UnreadRune", "", (go$funcType([], [go$error], false))]]);
		scanError.init([["err", "err", "fmt", go$error, ""]]);
		(go$ptrType(ss)).methods = [["Read", "", [(go$sliceType(Go$Uint8))], [Go$Int, go$error], false, -1], ["ReadRune", "", [], [Go$Int32, Go$Int, go$error], false, -1], ["SkipSpace", "", [], [], false, -1], ["Token", "", [Go$Bool, (go$funcType([Go$Int32], [Go$Bool], false))], [(go$sliceType(Go$Uint8)), go$error], false, -1], ["UnreadRune", "", [], [go$error], false, -1], ["Width", "", [], [Go$Int, Go$Bool], false, -1], ["accept", "fmt", [Go$String], [Go$Bool], false, -1], ["advance", "fmt", [Go$String], [Go$Int], false, -1], ["complexTokens", "fmt", [], [Go$String, Go$String], false, -1], ["consume", "fmt", [Go$String, Go$Bool], [Go$Bool], false, -1], ["convertFloat", "fmt", [Go$String, Go$Int], [Go$Float64], false, -1], ["convertString", "fmt", [Go$Int32], [Go$String], false, -1], ["doScan", "fmt", [(go$sliceType(go$emptyInterface))], [Go$Int, go$error], false, -1], ["doScanf", "fmt", [Go$String, (go$sliceType(go$emptyInterface))], [Go$Int, go$error], false, -1], ["error", "fmt", [go$error], [], false, -1], ["errorString", "fmt", [Go$String], [], false, -1], ["floatToken", "fmt", [], [Go$String], false, -1], ["free", "fmt", [ssave], [], false, -1], ["getBase", "fmt", [Go$Int32], [Go$Int, Go$String], false, -1], ["getRune", "fmt", [], [Go$Int32], false, -1], ["hexByte", "fmt", [], [Go$Uint8, Go$Bool], false, -1], ["hexDigit", "fmt", [Go$Int32], [Go$Int], false, -1], ["hexString", "fmt", [], [Go$String], false, -1], ["mustReadRune", "fmt", [], [Go$Int32], false, -1], ["notEOF", "fmt", [], [], false, -1], ["okVerb", "fmt", [Go$Int32, Go$String, Go$String], [Go$Bool], false, -1], ["peek", "fmt", [Go$String], [Go$Bool], false, -1], ["quotedString", "fmt", [], [Go$String], false, -1], ["scanBasePrefix", "fmt", [], [Go$Int, Go$String, Go$Bool], false, -1], ["scanBool", "fmt", [Go$Int32], [Go$Bool], false, -1], ["scanComplex", "fmt", [Go$Int32, Go$Int], [Go$Complex128], false, -1], ["scanInt", "fmt", [Go$Int32, Go$Int], [Go$Int64], false, -1], ["scanNumber", "fmt", [Go$String, Go$Bool], [Go$String], false, -1], ["scanOne", "fmt", [Go$Int32, go$emptyInterface], [], false, -1], ["scanRune", "fmt", [Go$Int], [Go$Int64], false, -1], ["scanUint", "fmt", [Go$Int32, Go$Int], [Go$Uint64], false, -1], ["skipSpace", "fmt", [Go$Bool], [], false, -1], ["token", "fmt", [Go$Bool, (go$funcType([Go$Int32], [Go$Bool], false))], [(go$sliceType(Go$Uint8))], false, -1]];
		ss.init([["rr", "rr", "fmt", io.RuneReader, ""], ["buf", "buf", "fmt", buffer, ""], ["peekRune", "peekRune", "fmt", Go$Int32, ""], ["prevRune", "prevRune", "fmt", Go$Int32, ""], ["count", "count", "fmt", Go$Int, ""], ["atEOF", "atEOF", "fmt", Go$Bool, ""], ["ssave", "", "fmt", ssave, ""]]);
		ssave.init([["validSave", "validSave", "fmt", Go$Bool, ""], ["nlIsEnd", "nlIsEnd", "fmt", Go$Bool, ""], ["nlIsSpace", "nlIsSpace", "fmt", Go$Bool, ""], ["argLimit", "argLimit", "fmt", Go$Int, ""], ["limit", "limit", "fmt", Go$Int, ""], ["maxWid", "maxWid", "fmt", Go$Int, ""]]);
		padZeroBytes = (go$sliceType(Go$Uint8)).make(65, 0, function() { return 0; });
		padSpaceBytes = (go$sliceType(Go$Uint8)).make(65, 0, function() { return 0; });
		trueBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("true"));
		falseBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("false"));
		commaSpaceBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes(", "));
		nilAngleBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("<nil>"));
		nilParenBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("(nil)"));
		nilBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("nil"));
		mapBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("map["));
		percentBangBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("%!"));
		missingBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("(MISSING)"));
		badIndexBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("(BADINDEX)"));
		panicBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("(PANIC="));
		extraBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("%!(EXTRA "));
		irparenBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("i)"));
		bytesBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("[]byte{"));
		badWidthBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("%!(BADWIDTH)"));
		badPrecBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("%!(BADPREC)"));
		noVerbBytes = new (go$sliceType(Go$Uint8))(go$stringToBytes("%!(NOVERB)"));
		ppFree = newCache((function() {
			return new pp.Ptr();
		}));
		intBits = reflect.TypeOf(new Go$Int(0)).Bits();
		uintptrBits = reflect.TypeOf(new Go$Uintptr(0)).Bits();
		space = new (go$sliceType((go$arrayType(Go$Uint16, 2))))([go$toNativeArray("Uint16", [9, 13]), go$toNativeArray("Uint16", [32, 32]), go$toNativeArray("Uint16", [133, 133]), go$toNativeArray("Uint16", [160, 160]), go$toNativeArray("Uint16", [5760, 5760]), go$toNativeArray("Uint16", [6158, 6158]), go$toNativeArray("Uint16", [8192, 8202]), go$toNativeArray("Uint16", [8232, 8233]), go$toNativeArray("Uint16", [8239, 8239]), go$toNativeArray("Uint16", [8287, 8287]), go$toNativeArray("Uint16", [12288, 12288])]);
		ssFree = newCache((function() {
			return new ss.Ptr();
		}));
		complexError = errors.New("syntax error scanning complex number");
		boolError = errors.New("syntax error scanning boolean");
		var i;
		i = 0;
		while (i < 65) {
			(i < 0 || i >= padZeroBytes.length) ? go$throwRuntimeError("index out of range") : padZeroBytes.array[padZeroBytes.offset + i] = 48;
			(i < 0 || i >= padSpaceBytes.length) ? go$throwRuntimeError("index out of range") : padSpaceBytes.array[padSpaceBytes.offset + i] = 32;
			i = i + 1 >> 0;
		}
	}
	return go$pkg;
})();
go$packages["todomvc/concorde"] = (function() {
	var go$pkg = {}, js = go$packages["github.com/gopherjs/gopherjs/js"], jquery = go$packages["github.com/gopherjs/jquery"], fmt = go$packages["fmt"], strings = go$packages["strings"], BooleanAttribute, BooleanSimple, BoolEqualer, BooleanInverter, BoolEq, builder, getDomAttrFunc, builderBase, anyBuilder, booleanBuilder, Collection, Joiner, EqList, computedEmpty, computedLength, Puller, FoldingIterator, foldedConstraint, Equaler, Attribute, Constraint, CssClass, cssClassImpl, htmlIdImpl, DomAttribute, domAttr, styleAttr, textAttr, htmlAttrAttr, propAttr, htmlAttrName, propName, cssExistenceAttr, edgeImpl, NodeType, node, AttributeImpl, ValueFunc, SideEffectFunc, EventFunc, EventName, eventHandler, IntegerAttribute, IntegerSimple, IntEqualer, StringAttribute, StringSimple, StringEqualer, eqConstraint, ModelName, Model, ModelImpl, ViewImpl, option, Main, NewBooleanSimple, NewBooleanInverter, textAttrBuilder, htmlAttrBuilder, propBuilder, cssExistenceBuilder, NewList, NewCssClass, NewHtmlId, newDomAttr, NewStyleAttr, NewTextAttr, NewHtmlAttrAttr, NewPropAttr, newAttrName, newPropName, NewCssExistenceAttr, NewDisplayAttr, DrainEagerQueue, newEdge, dropEdge, NewAttribute, NewIntegerSimple, NewStringSimple, Equality, NewModelName, Event, Class, ModelId, addBuilder, CssExistence, PropEqual, HtmlAttrEqual, TextEqual, HtmlIdFromModel, DIV, INPUT, LABEL, LI, BUTTON, tag, eagerQueue, numNodes, count;
	BooleanAttribute = go$pkg.BooleanAttribute = go$newType(8, "Interface", "concorde.BooleanAttribute", "BooleanAttribute", "todomvc/concorde", null);
	BooleanSimple = go$pkg.BooleanSimple = go$newType(0, "Struct", "concorde.BooleanSimple", "BooleanSimple", "todomvc/concorde", function(AttributeImpl_) {
		this.go$val = this;
		this.AttributeImpl = AttributeImpl_ !== undefined ? AttributeImpl_ : (go$ptrType(AttributeImpl)).nil;
	});
	BoolEqualer = go$pkg.BoolEqualer = go$newType(0, "Struct", "concorde.BoolEqualer", "BoolEqualer", "todomvc/concorde", function(B_) {
		this.go$val = this;
		this.B = B_ !== undefined ? B_ : false;
	});
	BooleanInverter = go$pkg.BooleanInverter = go$newType(0, "Struct", "concorde.BooleanInverter", "BooleanInverter", "todomvc/concorde", function(dep_) {
		this.go$val = this;
		this.dep = dep_ !== undefined ? dep_ : null;
	});
	BoolEq = go$pkg.BoolEq = go$newType(0, "Struct", "concorde.BoolEq", "BoolEq", "todomvc/concorde", function(dep_) {
		this.go$val = this;
		this.dep = dep_ !== undefined ? dep_ : null;
	});
	builder = go$pkg.builder = go$newType(8, "Interface", "concorde.builder", "builder", "todomvc/concorde", null);
	getDomAttrFunc = go$pkg.getDomAttrFunc = go$newType(4, "Func", "concorde.getDomAttrFunc", "getDomAttrFunc", "todomvc/concorde", null);
	builderBase = go$pkg.builderBase = go$newType(0, "Struct", "concorde.builderBase", "builderBase", "todomvc/concorde", function(cons_, get_) {
		this.go$val = this;
		this.cons = cons_ !== undefined ? cons_ : null;
		this.get = get_ !== undefined ? get_ : go$throwNilPointerError;
	});
	anyBuilder = go$pkg.anyBuilder = go$newType(0, "Struct", "concorde.anyBuilder", "anyBuilder", "todomvc/concorde", function(builderBase_, a_) {
		this.go$val = this;
		this.builderBase = builderBase_ !== undefined ? builderBase_ : new builderBase.Ptr();
		this.a = a_ !== undefined ? a_ : null;
	});
	booleanBuilder = go$pkg.booleanBuilder = go$newType(0, "Struct", "concorde.booleanBuilder", "booleanBuilder", "todomvc/concorde", function(builderBase_, b_) {
		this.go$val = this;
		this.builderBase = builderBase_ !== undefined ? builderBase_ : new builderBase.Ptr();
		this.b = b_ !== undefined ? b_ : null;
	});
	Collection = go$pkg.Collection = go$newType(0, "Struct", "concorde.Collection", "Collection", "todomvc/concorde", function(coll_, join_) {
		this.go$val = this;
		this.coll = coll_ !== undefined ? coll_ : (go$ptrType(AttributeImpl)).nil;
		this.join = join_ !== undefined ? join_ : (go$sliceType(Joiner)).nil;
	});
	Joiner = go$pkg.Joiner = go$newType(8, "Interface", "concorde.Joiner", "Joiner", "todomvc/concorde", null);
	EqList = go$pkg.EqList = go$newType(12, "Slice", "concorde.EqList", "EqList", "todomvc/concorde", null);
	computedEmpty = go$pkg.computedEmpty = go$newType(0, "Struct", "concorde.computedEmpty", "computedEmpty", "todomvc/concorde", function(AttributeImpl_, coll_) {
		this.go$val = this;
		this.AttributeImpl = AttributeImpl_ !== undefined ? AttributeImpl_ : (go$ptrType(AttributeImpl)).nil;
		this.coll = coll_ !== undefined ? coll_ : (go$ptrType(Collection)).nil;
	});
	computedLength = go$pkg.computedLength = go$newType(0, "Struct", "concorde.computedLength", "computedLength", "todomvc/concorde", function(AttributeImpl_, coll_) {
		this.go$val = this;
		this.AttributeImpl = AttributeImpl_ !== undefined ? AttributeImpl_ : (go$ptrType(AttributeImpl)).nil;
		this.coll = coll_ !== undefined ? coll_ : (go$ptrType(Collection)).nil;
	});
	Puller = go$pkg.Puller = go$newType(4, "Func", "concorde.Puller", "Puller", "todomvc/concorde", null);
	FoldingIterator = go$pkg.FoldingIterator = go$newType(4, "Func", "concorde.FoldingIterator", "FoldingIterator", "todomvc/concorde", null);
	foldedConstraint = go$pkg.foldedConstraint = go$newType(0, "Struct", "concorde.foldedConstraint", "foldedConstraint", "todomvc/concorde", function(deps_, fn_, pull_, initial_, targ_, empty_) {
		this.go$val = this;
		this.deps = deps_ !== undefined ? deps_ : (go$sliceType(Attribute)).nil;
		this.fn = fn_ !== undefined ? fn_ : go$throwNilPointerError;
		this.pull = pull_ !== undefined ? pull_ : go$throwNilPointerError;
		this.initial = initial_ !== undefined ? initial_ : null;
		this.targ = targ_ !== undefined ? targ_ : null;
		this.empty = empty_ !== undefined ? empty_ : null;
	});
	Equaler = go$pkg.Equaler = go$newType(8, "Interface", "concorde.Equaler", "Equaler", "todomvc/concorde", null);
	Attribute = go$pkg.Attribute = go$newType(8, "Interface", "concorde.Attribute", "Attribute", "todomvc/concorde", null);
	Constraint = go$pkg.Constraint = go$newType(8, "Interface", "concorde.Constraint", "Constraint", "todomvc/concorde", null);
	CssClass = go$pkg.CssClass = go$newType(8, "Interface", "concorde.CssClass", "CssClass", "todomvc/concorde", null);
	cssClassImpl = go$pkg.cssClassImpl = go$newType(0, "Struct", "concorde.cssClassImpl", "cssClassImpl", "todomvc/concorde", function(name_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : "";
	});
	htmlIdImpl = go$pkg.htmlIdImpl = go$newType(0, "Struct", "concorde.htmlIdImpl", "htmlIdImpl", "todomvc/concorde", function(tag_, id_) {
		this.go$val = this;
		this.tag = tag_ !== undefined ? tag_ : "";
		this.id = id_ !== undefined ? id_ : "";
	});
	DomAttribute = go$pkg.DomAttribute = go$newType(8, "Interface", "concorde.DomAttribute", "DomAttribute", "todomvc/concorde", null);
	domAttr = go$pkg.domAttr = go$newType(0, "Struct", "concorde.domAttr", "domAttr", "todomvc/concorde", function(attr_, j_, id_, set_, get_) {
		this.go$val = this;
		this.attr = attr_ !== undefined ? attr_ : (go$ptrType(AttributeImpl)).nil;
		this.j = j_ !== undefined ? j_ : new jquery.JQuery.Ptr();
		this.id = id_ !== undefined ? id_ : "";
		this.set = set_ !== undefined ? set_ : go$throwNilPointerError;
		this.get = get_ !== undefined ? get_ : go$throwNilPointerError;
	});
	styleAttr = go$pkg.styleAttr = go$newType(0, "Struct", "concorde.styleAttr", "styleAttr", "todomvc/concorde", function(domAttr_, name_) {
		this.go$val = this;
		this.domAttr = domAttr_ !== undefined ? domAttr_ : (go$ptrType(domAttr)).nil;
		this.name = name_ !== undefined ? name_ : "";
	});
	textAttr = go$pkg.textAttr = go$newType(0, "Struct", "concorde.textAttr", "textAttr", "todomvc/concorde", function(domAttr_) {
		this.go$val = this;
		this.domAttr = domAttr_ !== undefined ? domAttr_ : (go$ptrType(domAttr)).nil;
	});
	htmlAttrAttr = go$pkg.htmlAttrAttr = go$newType(0, "Struct", "concorde.htmlAttrAttr", "htmlAttrAttr", "todomvc/concorde", function(domAttr_, name_) {
		this.go$val = this;
		this.domAttr = domAttr_ !== undefined ? domAttr_ : (go$ptrType(domAttr)).nil;
		this.name = name_ !== undefined ? name_ : "";
	});
	propAttr = go$pkg.propAttr = go$newType(0, "Struct", "concorde.propAttr", "propAttr", "todomvc/concorde", function(domAttr_, name_) {
		this.go$val = this;
		this.domAttr = domAttr_ !== undefined ? domAttr_ : (go$ptrType(domAttr)).nil;
		this.name = name_ !== undefined ? name_ : "";
	});
	htmlAttrName = go$pkg.htmlAttrName = go$newType(8, "String", "concorde.htmlAttrName", "htmlAttrName", "todomvc/concorde", null);
	propName = go$pkg.propName = go$newType(8, "String", "concorde.propName", "propName", "todomvc/concorde", null);
	cssExistenceAttr = go$pkg.cssExistenceAttr = go$newType(0, "Struct", "concorde.cssExistenceAttr", "cssExistenceAttr", "todomvc/concorde", function(domAttr_, clazz_) {
		this.go$val = this;
		this.domAttr = domAttr_ !== undefined ? domAttr_ : (go$ptrType(domAttr)).nil;
		this.clazz = clazz_ !== undefined ? clazz_ : null;
	});
	edgeImpl = go$pkg.edgeImpl = go$newType(0, "Struct", "concorde.edgeImpl", "edgeImpl", "todomvc/concorde", function(notMarked_, d_, s_) {
		this.go$val = this;
		this.notMarked = notMarked_ !== undefined ? notMarked_ : false;
		this.d = d_ !== undefined ? d_ : null;
		this.s = s_ !== undefined ? s_ : null;
	});
	NodeType = go$pkg.NodeType = go$newType(4, "Int", "concorde.NodeType", "NodeType", "todomvc/concorde", null);
	node = go$pkg.node = go$newType(8, "Interface", "concorde.node", "node", "todomvc/concorde", null);
	AttributeImpl = go$pkg.AttributeImpl = go$newType(0, "Struct", "concorde.AttributeImpl", "AttributeImpl", "todomvc/concorde", function(n_, evalCount_, demandCount_, clean_, edge_, constraint_, curr_, sideEffectFn_, valueFn_, nType_) {
		this.go$val = this;
		this.n = n_ !== undefined ? n_ : 0;
		this.evalCount = evalCount_ !== undefined ? evalCount_ : 0;
		this.demandCount = demandCount_ !== undefined ? demandCount_ : 0;
		this.clean = clean_ !== undefined ? clean_ : false;
		this.edge = edge_ !== undefined ? edge_ : (go$sliceType((go$ptrType(edgeImpl)))).nil;
		this.constraint = constraint_ !== undefined ? constraint_ : null;
		this.curr = curr_ !== undefined ? curr_ : null;
		this.sideEffectFn = sideEffectFn_ !== undefined ? sideEffectFn_ : go$throwNilPointerError;
		this.valueFn = valueFn_ !== undefined ? valueFn_ : go$throwNilPointerError;
		this.nType = nType_ !== undefined ? nType_ : 0;
	});
	ValueFunc = go$pkg.ValueFunc = go$newType(4, "Func", "concorde.ValueFunc", "ValueFunc", "todomvc/concorde", null);
	SideEffectFunc = go$pkg.SideEffectFunc = go$newType(4, "Func", "concorde.SideEffectFunc", "SideEffectFunc", "todomvc/concorde", null);
	EventFunc = go$pkg.EventFunc = go$newType(4, "Func", "concorde.EventFunc", "EventFunc", "todomvc/concorde", null);
	EventName = go$pkg.EventName = go$newType(4, "Int", "concorde.EventName", "EventName", "todomvc/concorde", null);
	eventHandler = go$pkg.eventHandler = go$newType(0, "Struct", "concorde.eventHandler", "eventHandler", "todomvc/concorde", function(name_, j_, fn_) {
		this.go$val = this;
		this.name = name_ !== undefined ? name_ : 0;
		this.j = j_ !== undefined ? j_ : new jquery.JQuery.Ptr();
		this.fn = fn_ !== undefined ? fn_ : go$throwNilPointerError;
	});
	IntegerAttribute = go$pkg.IntegerAttribute = go$newType(8, "Interface", "concorde.IntegerAttribute", "IntegerAttribute", "todomvc/concorde", null);
	IntegerSimple = go$pkg.IntegerSimple = go$newType(0, "Struct", "concorde.IntegerSimple", "IntegerSimple", "todomvc/concorde", function(AttributeImpl_) {
		this.go$val = this;
		this.AttributeImpl = AttributeImpl_ !== undefined ? AttributeImpl_ : (go$ptrType(AttributeImpl)).nil;
	});
	IntEqualer = go$pkg.IntEqualer = go$newType(0, "Struct", "concorde.IntEqualer", "IntEqualer", "todomvc/concorde", function(I_) {
		this.go$val = this;
		this.I = I_ !== undefined ? I_ : 0;
	});
	StringAttribute = go$pkg.StringAttribute = go$newType(8, "Interface", "concorde.StringAttribute", "StringAttribute", "todomvc/concorde", null);
	StringSimple = go$pkg.StringSimple = go$newType(0, "Struct", "concorde.StringSimple", "StringSimple", "todomvc/concorde", function(AttributeImpl_) {
		this.go$val = this;
		this.AttributeImpl = AttributeImpl_ !== undefined ? AttributeImpl_ : (go$ptrType(AttributeImpl)).nil;
	});
	StringEqualer = go$pkg.StringEqualer = go$newType(0, "Struct", "concorde.StringEqualer", "StringEqualer", "todomvc/concorde", function(S_) {
		this.go$val = this;
		this.S = S_ !== undefined ? S_ : "";
	});
	eqConstraint = go$pkg.eqConstraint = go$newType(0, "Struct", "concorde.eqConstraint", "eqConstraint", "todomvc/concorde", function(dep_) {
		this.go$val = this;
		this.dep = dep_ !== undefined ? dep_ : null;
	});
	ModelName = go$pkg.ModelName = go$newType(8, "Interface", "concorde.ModelName", "ModelName", "todomvc/concorde", null);
	Model = go$pkg.Model = go$newType(8, "Interface", "concorde.Model", "Model", "todomvc/concorde", null);
	ModelImpl = go$pkg.ModelImpl = go$newType(0, "Struct", "concorde.ModelImpl", "ModelImpl", "todomvc/concorde", function(n_, t_) {
		this.go$val = this;
		this.n = n_ !== undefined ? n_ : 0;
		this.t = t_ !== undefined ? t_ : "";
	});
	ViewImpl = go$pkg.ViewImpl = go$newType(0, "Struct", "concorde.ViewImpl", "ViewImpl", "todomvc/concorde", function(tag_, id_, classes_, children_, text_, builders_, event_) {
		this.go$val = this;
		this.tag = tag_ !== undefined ? tag_ : "";
		this.id = id_ !== undefined ? id_ : "";
		this.classes = classes_ !== undefined ? classes_ : (go$sliceType(Go$String)).nil;
		this.children = children_ !== undefined ? children_ : (go$sliceType((go$ptrType(ViewImpl)))).nil;
		this.text = text_ !== undefined ? text_ : "";
		this.builders = builders_ !== undefined ? builders_ : (go$sliceType(builder)).nil;
		this.event = event_ !== undefined ? event_ : (go$sliceType((go$ptrType(eventHandler)))).nil;
	});
	option = go$pkg.option = go$newType(4, "Func", "concorde.option", "option", "todomvc/concorde", null);
	Main = go$pkg.Main = function(app) {
		jquery.NewJQuery(new (go$sliceType(go$emptyInterface))([go$global.document])).Ready((function() {
			app.Start();
		}));
	};
	BoolEqualer.Ptr.prototype.Equal = function(e) {
		var _struct, self;
		self = (_struct = this, new BoolEqualer.Ptr(_struct.B));
		return self.B === (e !== null && e.constructor === BoolEqualer ? e.go$val : go$typeAssertionFailed(e, BoolEqualer)).B;
	};
	BoolEqualer.prototype.Equal = function(e) { return this.go$val.Equal(e); };
	BoolEqualer.Ptr.prototype.String = function() {
		var _struct, self;
		self = (_struct = this, new BoolEqualer.Ptr(_struct.B));
		return fmt.Sprintf("%v", new (go$sliceType(go$emptyInterface))([new Go$Bool(self.B)]));
	};
	BoolEqualer.prototype.String = function() { return this.go$val.String(); };
	BooleanSimple.Ptr.prototype.Value = function() {
		var self, x;
		self = this;
		return (x = self.AttributeImpl.Demand(), (x !== null && x.constructor === BoolEqualer ? x.go$val : go$typeAssertionFailed(x, BoolEqualer))).B;
	};
	BooleanSimple.prototype.Value = function() { return this.go$val.Value(); };
	BooleanSimple.Ptr.prototype.Set = function(b) {
		var self, x;
		self = this;
		self.AttributeImpl.SetEqualer((x = new BoolEqualer.Ptr(b), new x.constructor.Struct(x)));
	};
	BooleanSimple.prototype.Set = function(b) { return this.go$val.Set(b); };
	NewBooleanSimple = go$pkg.NewBooleanSimple = function(b) {
		var result;
		result = new BooleanSimple.Ptr(NewAttribute(0, go$throwNilPointerError, go$throwNilPointerError));
		result.Set(b);
		return result;
	};
	BooleanInverter.Ptr.prototype.Inputs = function() {
		var self;
		self = this;
		return new (go$sliceType(Attribute))([self.dep]);
	};
	BooleanInverter.prototype.Inputs = function() { return this.go$val.Inputs(); };
	BooleanInverter.Ptr.prototype.Fn = function(in$1) {
		var self, x, x$1;
		self = this;
		if (!((in$1.length === 1))) {
			throw go$panic(new Go$String("unexpected number of parameters to boolean inverter!"));
		}
		return (x = new BoolEqualer.Ptr(!(x$1 = ((0 < 0 || 0 >= in$1.length) ? go$throwRuntimeError("index out of range") : in$1.array[in$1.offset + 0]), (x$1 !== null && x$1.constructor === BoolEqualer ? x$1.go$val : go$typeAssertionFailed(x$1, BoolEqualer))).B), new x.constructor.Struct(x));
	};
	BooleanInverter.prototype.Fn = function(in$1) { return this.go$val.Fn(in$1); };
	NewBooleanInverter = go$pkg.NewBooleanInverter = function(attr) {
		var result;
		result = new BooleanInverter.Ptr(attr);
		return result;
	};
	BoolEq.Ptr.prototype.Inputs = function() {
		var self;
		self = this;
		return new (go$sliceType(Attribute))([self.dep]);
	};
	BoolEq.prototype.Inputs = function() { return this.go$val.Inputs(); };
	BoolEq.Ptr.prototype.Fn = function(in$1) {
		var self, x;
		self = this;
		return (x = new BoolEqualer.Ptr(self.dep.Value()), new x.constructor.Struct(x));
	};
	BoolEq.prototype.Fn = function(in$1) { return this.go$val.Fn(in$1); };
	anyBuilder.Ptr.prototype.build = function(j) {
		var self, _struct, dest, x;
		self = this;
		dest = self.builderBase.get((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)));
		if (go$interfaceIsEqual(self.builderBase.cons, null)) {
			self.builderBase.cons = (x = new eqConstraint.Ptr(self.a), new x.constructor.Struct(x));
		}
		dest.Attach(self.builderBase.cons);
	};
	anyBuilder.prototype.build = function(j) { return this.go$val.build(j); };
	booleanBuilder.Ptr.prototype.build = function(j) {
		var self, _struct, dest;
		self = this;
		dest = self.builderBase.get((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)));
		if (go$interfaceIsEqual(self.builderBase.cons, null)) {
			self.builderBase.cons = new BoolEq.Ptr(self.b);
		}
		dest.Attach(self.builderBase.cons);
	};
	booleanBuilder.prototype.build = function(j) { return this.go$val.build(j); };
	textAttrBuilder = function(attr, cons) {
		return new anyBuilder.Ptr(new builderBase.Ptr(cons, (function(j) {
			var _struct;
			return NewTextAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)));
		})), attr);
	};
	htmlAttrBuilder = function(h, attr, cons) {
		return new anyBuilder.Ptr(new builderBase.Ptr(cons, (function(j) {
			var _struct;
			return NewHtmlAttrAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), h);
		})), attr);
	};
	propBuilder = function(p, attr, cons) {
		return new booleanBuilder.Ptr(new builderBase.Ptr(cons, (function(j) {
			var _struct;
			return NewPropAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), p);
		})), attr);
	};
	cssExistenceBuilder = function(c, b) {
		return new booleanBuilder.Ptr(new builderBase.Ptr(null, (function(j) {
			var _struct;
			return NewCssExistenceAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), c);
		})), b);
	};
	EqList.prototype.Equal = function(o) {
		var self, other, curr, _ref, _i, e, i;
		self = this;
		if (go$interfaceIsEqual(o, null)) {
			return false;
		}
		other = (o !== null && o.constructor === EqList ? o.go$val : go$typeAssertionFailed(o, EqList));
		if (other.length === 0) {
			return false;
		}
		curr = go$subslice(new (go$sliceType(Model))(self.array), self.offset, self.offset + self.length);
		if (curr.length === 0) {
			return false;
		}
		if (!((curr.length === other.length))) {
			return false;
		}
		_ref = other;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if (!((i < 0 || i >= curr.length) ? go$throwRuntimeError("index out of range") : curr.array[curr.offset + i]).Equal(e)) {
				return false;
			}
			_i++;
		}
		return true;
	};
	go$ptrType(EqList).prototype.Equal = function(o) { return this.go$get().Equal(o); };
	Collection.Ptr.prototype.empty = function() {
		var self, x, x$1, x$2;
		self = this;
		if (go$interfaceIsEqual(self.coll.Demand(), null) || ((x = self.coll.Demand(), (x !== null && x.constructor === EqList ? x.go$val : go$typeAssertionFailed(x, EqList))).length === 0)) {
			return (x$1 = new BoolEqualer.Ptr(true), new x$1.constructor.Struct(x$1));
		}
		return (x$2 = new BoolEqualer.Ptr(false), new x$2.constructor.Struct(x$2));
	};
	Collection.prototype.empty = function() { return this.go$val.empty(); };
	computedEmpty.Ptr.prototype.Set = function(b) {
		var self;
		self = this;
		throw go$panic(new Go$String("can't set the value of a computed attribute (empty of the list)"));
	};
	computedEmpty.prototype.Set = function(b) { return this.go$val.Set(b); };
	computedEmpty.Ptr.prototype.Value = function() {
		var self, x;
		self = this;
		self.AttributeImpl.clean = true;
		return (x = self.coll.empty(), (x !== null && x.constructor === BoolEqualer ? x.go$val : go$typeAssertionFailed(x, BoolEqualer))).B;
	};
	computedEmpty.prototype.Value = function() { return this.go$val.Value(); };
	Collection.Ptr.prototype.EmptyAttribute = function() {
		var self, _recv, result;
		self = this;
		result = new computedEmpty.Ptr(NewAttribute(1, (_recv = self, function() { return _recv.empty(); }), go$throwNilPointerError), self);
		newEdge(self.coll, result);
		return result;
	};
	Collection.prototype.EmptyAttribute = function() { return this.go$val.EmptyAttribute(); };
	computedLength.Ptr.prototype.Set = function(i) {
		var self;
		self = this;
		throw go$panic(new Go$String("can't set the value of a computed attribute (length)"));
	};
	computedLength.prototype.Set = function(i) { return this.go$val.Set(i); };
	computedLength.Ptr.prototype.Value = function() {
		var self, x;
		self = this;
		return (x = self.coll.length(), (x !== null && x.constructor === IntEqualer ? x.go$val : go$typeAssertionFailed(x, IntEqualer))).I;
	};
	computedLength.prototype.Value = function() { return this.go$val.Value(); };
	Collection.Ptr.prototype.length = function() {
		var self, x, x$1;
		self = this;
		return (x = new IntEqualer.Ptr((x$1 = self.coll.Demand(), (x$1 !== null && x$1.constructor === EqList ? x$1.go$val : go$typeAssertionFailed(x$1, EqList))).length), new x.constructor.Struct(x));
	};
	Collection.prototype.length = function() { return this.go$val.length(); };
	Collection.Ptr.prototype.LengthAttribute = function() {
		var self, _recv, result;
		self = this;
		result = new computedLength.Ptr(NewAttribute(1, (_recv = self, function() { return _recv.length(); }), go$throwNilPointerError), self);
		return result;
	};
	Collection.prototype.LengthAttribute = function() { return this.go$val.LengthAttribute(); };
	Collection.Ptr.prototype.PushRaw = function(m) {
		var self, current, result, _ref, _i, j;
		self = this;
		current = self.coll.Demand();
		result = EqList.nil;
		if (go$interfaceIsEqual(current, null)) {
			result = new EqList([m]);
		} else {
			result = go$append((current !== null && current.constructor === EqList ? current.go$val : go$typeAssertionFailed(current, EqList)), m);
		}
		self.coll.SetEqualer(result);
		if (!(self.join === (go$sliceType(Joiner)).nil)) {
			_ref = self.join;
			_i = 0;
			while (_i < _ref.length) {
				j = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
				j.Add(result.length, m);
				_i++;
			}
		}
	};
	Collection.prototype.PushRaw = function(m) { return this.go$val.PushRaw(m); };
	Collection.Ptr.prototype.PopRaw = function() {
		var self, x, x$1, obj, x$2, result, _ref, _i, j;
		self = this;
		if (go$interfaceIsEqual(self.coll.Demand(), null) || ((x = self.coll.Demand(), (x !== null && x.constructor === EqList ? x.go$val : go$typeAssertionFailed(x, EqList))).length === 0)) {
			throw go$panic(new Go$String("can't pop from a empty ListNode!"));
		}
		obj = (x$1 = self.coll.Demand(), (x$1 !== null && x$1.constructor === EqList ? x$1.go$val : go$typeAssertionFailed(x$1, EqList)));
		self.coll.markDirty();
		result = (x$2 = obj.length - 1 >> 0, ((x$2 < 0 || x$2 >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + x$2]));
		if (obj.length === 1) {
			self.coll.SetEqualer(null);
		} else {
			self.coll.SetEqualer(go$subslice(obj, 0, (obj.length - 1 >> 0)));
		}
		if (!(self.join === (go$sliceType(Joiner)).nil)) {
			_ref = self.join;
			_i = 0;
			while (_i < _ref.length) {
				j = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
				j.Remove(obj.length - 1 >> 0, result);
				_i++;
			}
		}
		return result;
	};
	Collection.prototype.PopRaw = function() { return this.go$val.PopRaw(); };
	NewList = go$pkg.NewList = function(joiner) {
		var result;
		result = new Collection.Ptr(NewAttribute(1, go$throwNilPointerError, go$throwNilPointerError), (go$sliceType(Joiner)).nil);
		if (!(go$interfaceIsEqual(joiner, null))) {
			result.join = new (go$sliceType(Joiner))([joiner]);
		}
		return result;
	};
	Collection.Ptr.prototype.Add = function(m) {
		var self, x, obj, _ref, _i, e;
		self = this;
		obj = (x = self.coll.Demand(), (x !== null && x.constructor === EqList ? x.go$val : go$typeAssertionFailed(x, EqList)));
		_ref = obj;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			if (e.Equal(m)) {
				return;
			}
			_i++;
		}
		self.PushRaw(m);
	};
	Collection.prototype.Add = function(m) { return this.go$val.Add(m); };
	Collection.Ptr.prototype.Remove = function(m) {
		var self, x, obj, _ref, _i, e, i, x$1, _ref$1, _i$1, j;
		self = this;
		obj = (x = self.coll.Demand(), (x !== null && x.constructor === EqList ? x.go$val : go$typeAssertionFailed(x, EqList)));
		_ref = obj;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if (e.Equal(m)) {
				if (i === (obj.length - 1 >> 0)) {
					self.PopRaw();
				} else {
					go$copySlice(go$subslice(obj, i), go$subslice(obj, (i + 1 >> 0)));
					(x$1 = obj.length - 1 >> 0, (x$1 < 0 || x$1 >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + x$1] = null);
					self.coll.SetEqualer(go$subslice(obj, 0, (obj.length - 1 >> 0)));
					if (!(self.join === (go$sliceType(Joiner)).nil)) {
						_ref$1 = self.join;
						_i$1 = 0;
						while (_i$1 < _ref$1.length) {
							j = ((_i$1 < 0 || _i$1 >= _ref$1.length) ? go$throwRuntimeError("index out of range") : _ref$1.array[_ref$1.offset + _i$1]);
							j.Remove(i, m);
							_i$1++;
						}
					}
				}
				break;
			}
			_i++;
		}
	};
	Collection.prototype.Remove = function(m) { return this.go$val.Remove(m); };
	Collection.Ptr.prototype.All = function() {
		var self, x, curr, result, _ref, _i, c, i;
		self = this;
		if (go$interfaceIsEqual(self.coll.Demand(), null)) {
			return (go$sliceType(Model)).nil;
		}
		curr = (x = self.coll.Demand(), (x !== null && x.constructor === EqList ? x.go$val : go$typeAssertionFailed(x, EqList)));
		result = (go$sliceType(Model)).make(curr.length, 0, function() { return null; });
		_ref = curr;
		_i = 0;
		while (_i < _ref.length) {
			c = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			(i < 0 || i >= result.length) ? go$throwRuntimeError("index out of range") : result.array[result.offset + i] = c;
			_i++;
		}
		return result;
	};
	Collection.prototype.All = function() { return this.go$val.All(); };
	Collection.Ptr.prototype.AllFold = function(targ, initial, puller, folder, empty) {
		var self, result;
		self = this;
		result = new foldedConstraint.Ptr((go$sliceType(Attribute)).nil, folder, puller, initial, targ, empty);
		self.join = go$append(self.join, result);
		targ.Attach(result);
		return result;
	};
	Collection.prototype.AllFold = function(targ, initial, puller, folder, empty) { return this.go$val.AllFold(targ, initial, puller, folder, empty); };
	foldedConstraint.Ptr.prototype.Inputs = function() {
		var self;
		self = this;
		return self.deps;
	};
	foldedConstraint.prototype.Inputs = function() { return this.go$val.Inputs(); };
	foldedConstraint.Ptr.prototype.Fn = function(in$1) {
		var self, prev, e, _ref, _i, i, _tuple;
		self = this;
		if (self.deps.length === 0) {
			return self.empty;
		}
		prev = self.initial;
		e = null;
		_ref = in$1;
		_i = 0;
		while (_i < _ref.length) {
			i = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			_tuple = self.fn(prev, i); prev = _tuple[0]; e = _tuple[1];
			_i++;
		}
		return e;
	};
	foldedConstraint.prototype.Fn = function(in$1) { return this.go$val.Fn(in$1); };
	foldedConstraint.Ptr.prototype.Remove = function(i, m) {
		var self, x, x$1, x$2, x$3;
		self = this;
		dropEdge((x = (x$1 = self.deps, ((i < 0 || i >= x$1.length) ? go$throwRuntimeError("index out of range") : x$1.array[x$1.offset + i])), (x !== null && node.implementedBy.indexOf(x.constructor) !== -1 ? x : go$typeAssertionFailed(x, node))), (x$2 = self.targ, (x$2 !== null && node.implementedBy.indexOf(x$2.constructor) !== -1 ? x$2 : go$typeAssertionFailed(x$2, node))));
		(x$3 = self.targ, (x$3 !== null && node.implementedBy.indexOf(x$3.constructor) !== -1 ? x$3 : go$typeAssertionFailed(x$3, node))).markDirty();
		if ((self.deps.length - 1 >> 0) === i) {
			self.deps = go$subslice(self.deps, 0, (self.deps.length - 1 >> 0));
		} else {
			self.deps = go$appendSlice(go$subslice(self.deps, 0, i), go$subslice(self.deps, (i + 1 >> 0)));
		}
		DrainEagerQueue();
	};
	foldedConstraint.prototype.Remove = function(i, m) { return this.go$val.Remove(i, m); };
	foldedConstraint.Ptr.prototype.Add = function(i, m) {
		var self, a, x, x$1;
		self = this;
		a = self.pull(m);
		self.deps = go$append(self.deps, a);
		newEdge((a !== null && node.implementedBy.indexOf(a.constructor) !== -1 ? a : go$typeAssertionFailed(a, node)), (x = self.targ, (x !== null && node.implementedBy.indexOf(x.constructor) !== -1 ? x : go$typeAssertionFailed(x, node))));
		(x$1 = self.targ, (x$1 !== null && node.implementedBy.indexOf(x$1.constructor) !== -1 ? x$1 : go$typeAssertionFailed(x$1, node))).markDirty();
		DrainEagerQueue();
	};
	foldedConstraint.prototype.Add = function(i, m) { return this.go$val.Add(i, m); };
	cssClassImpl.Ptr.prototype.ClassName = function() {
		var _struct, self;
		self = (_struct = this, new cssClassImpl.Ptr(_struct.name));
		return self.name;
	};
	cssClassImpl.prototype.ClassName = function() { return this.go$val.ClassName(); };
	NewCssClass = go$pkg.NewCssClass = function(name) {
		var x;
		return (x = new cssClassImpl.Ptr(name), new x.constructor.Struct(x));
	};
	NewHtmlId = go$pkg.NewHtmlId = function(tag$1, id) {
		return new htmlIdImpl.Ptr(tag$1, id);
	};
	htmlIdImpl.Ptr.prototype.TagName = function() {
		var self;
		self = this;
		return self.tag;
	};
	htmlIdImpl.prototype.TagName = function() { return this.go$val.TagName(); };
	htmlIdImpl.Ptr.prototype.Id = function() {
		var self;
		self = this;
		return self.id;
	};
	htmlIdImpl.prototype.Id = function() { return this.go$val.Id(); };
	htmlIdImpl.Ptr.prototype.Select = function() {
		var self, _struct;
		self = this;
		return (_struct = jquery.NewJQuery(new (go$sliceType(go$emptyInterface))([new Go$String(self.TagName() + "#" + self.Id())])), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
	};
	htmlIdImpl.prototype.Select = function() { return this.go$val.Select(); };
	htmlIdImpl.Ptr.prototype.StyleAttribute = function(name) {
		var self, _struct;
		self = this;
		return NewStyleAttr(name, (_struct = self.Select(), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)));
	};
	htmlIdImpl.prototype.StyleAttribute = function(name) { return this.go$val.StyleAttribute(name); };
	htmlIdImpl.Ptr.prototype.DisplayAttribute = function() {
		var self, _struct;
		self = this;
		return NewDisplayAttr((_struct = self.Select(), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)));
	};
	htmlIdImpl.prototype.DisplayAttribute = function() { return this.go$val.DisplayAttribute(); };
	htmlIdImpl.Ptr.prototype.TextAttribute = function() {
		var self, _struct;
		self = this;
		return NewTextAttr((_struct = self.Select(), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)));
	};
	htmlIdImpl.prototype.TextAttribute = function() { return this.go$val.TextAttribute(); };
	htmlIdImpl.Ptr.prototype.Event = function(n, f) {
		var self, _struct, h;
		self = this;
		h = new eventHandler.Ptr(n, (_struct = self.Select(), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), f);
		h.register();
	};
	htmlIdImpl.prototype.Event = function(n, f) { return this.go$val.Event(n, f); };
	newDomAttr = function(j, id, g, s) {
		var _struct, result;
		result = new domAttr.Ptr((go$ptrType(AttributeImpl)).nil, (_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), id, go$throwNilPointerError, go$throwNilPointerError);
		result.attr = NewAttribute(2, g, s);
		return result;
	};
	domAttr.Ptr.prototype.getData = function() {
		var self, d;
		self = this;
		d = self.j.Data(fmt.Sprintf("concorde_%s", new (go$sliceType(go$emptyInterface))([new Go$String(self.id)])));
		if (go$interfaceIsEqual(d, null)) {
			return "";
		}
		return (d !== null && d.constructor === Go$String ? d.go$val : go$typeAssertionFailed(d, Go$String));
	};
	domAttr.prototype.getData = function() { return this.go$val.getData(); };
	domAttr.Ptr.prototype.removeData = function() {
		var self;
		self = this;
		self.j.SetData(fmt.Sprintf("concorde_%s", new (go$sliceType(go$emptyInterface))([new Go$String(self.id)])), null);
	};
	domAttr.prototype.removeData = function() { return this.go$val.removeData(); };
	domAttr.Ptr.prototype.setData = function(v) {
		var self;
		self = this;
		self.j.SetData(fmt.Sprintf("concorde_%s", new (go$sliceType(go$emptyInterface))([new Go$String(self.id)])), new Go$String(v));
	};
	domAttr.prototype.setData = function(v) { return this.go$val.setData(v); };
	domAttr.Ptr.prototype.verifyConstraint = function(b) {
		var self, s;
		self = this;
		s = self.getData();
		if (b) {
			if (s === "") {
				throw go$panic(new Go$String(fmt.Sprintf("expected to have constraint on %s but did not!", new (go$sliceType(go$emptyInterface))([new Go$String(self.id)]))));
			}
			if (!(s === "constraint")) {
				throw go$panic(new Go$String(fmt.Sprintf("expected to find constraint marker but found %s", new (go$sliceType(go$emptyInterface))([new Go$String(s)]))));
			}
		} else {
			if (!(s === "")) {
				throw go$panic(new Go$String(fmt.Sprintf("did not expect to have constraint on %s but found one: %s", new (go$sliceType(go$emptyInterface))([new Go$String(self.id), new Go$String(s)]))));
			}
		}
	};
	domAttr.prototype.verifyConstraint = function(b) { return this.go$val.verifyConstraint(b); };
	domAttr.Ptr.prototype.Attach = function(c) {
		var self;
		self = this;
		self.verifyConstraint(false);
		self.attr.Attach(c);
		self.setData("constraint");
	};
	domAttr.prototype.Attach = function(c) { return this.go$val.Attach(c); };
	domAttr.Ptr.prototype.Detach = function() {
		var self;
		self = this;
		self.verifyConstraint(true);
		self.attr.Detach();
		self.removeData();
	};
	domAttr.prototype.Detach = function() { return this.go$val.Detach(); };
	domAttr.Ptr.prototype.Demand = function() {
		var self;
		self = this;
		return self.attr.Demand();
	};
	domAttr.prototype.Demand = function() { return this.go$val.Demand(); };
	domAttr.Ptr.prototype.SetEqualer = function(e) {
		var self;
		self = this;
		self.attr.SetEqualer(e);
	};
	domAttr.prototype.SetEqualer = function(e) { return this.go$val.SetEqualer(e); };
	domAttr.Ptr.prototype.Id = function() {
		var self;
		self = this;
		return self.id;
	};
	domAttr.prototype.Id = function() { return this.go$val.Id(); };
	NewStyleAttr = go$pkg.NewStyleAttr = function(n, j) {
		var result, _struct, _recv, e, _recv$1;
		result = new styleAttr.Ptr((go$ptrType(domAttr)).nil, n);
		result.domAttr = newDomAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), fmt.Sprintf("style:%s", new (go$sliceType(go$emptyInterface))([new Go$String(n)])), (_recv = result, function() { return _recv.get(); }), (_recv$1 = result, function(e) { return _recv$1.set(e); }));
		return result;
	};
	styleAttr.Ptr.prototype.get = function() {
		var self, x, x$1;
		self = this;
		if (self.domAttr.j.Css(self.name) === "undefined") {
			return (x = new StringEqualer.Ptr(""), new x.constructor.Struct(x));
		}
		return (x$1 = new StringEqualer.Ptr(self.domAttr.j.Css(self.name)), new x$1.constructor.Struct(x$1));
	};
	styleAttr.prototype.get = function() { return this.go$val.get(); };
	styleAttr.Ptr.prototype.set = function(e) {
		var self;
		self = this;
		self.domAttr.j.SetCss(new (go$sliceType(go$emptyInterface))([new Go$String(self.name), new Go$String(fmt.Sprintf("%s", new (go$sliceType(go$emptyInterface))([e])))]));
	};
	styleAttr.prototype.set = function(e) { return this.go$val.set(e); };
	NewTextAttr = go$pkg.NewTextAttr = function(j) {
		var result, _struct, _recv, e, _recv$1;
		result = new textAttr.Ptr((go$ptrType(domAttr)).nil);
		result.domAttr = newDomAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), "text", (_recv = result, function() { return _recv.get(); }), (_recv$1 = result, function(e) { return _recv$1.set(e); }));
		return result;
	};
	textAttr.Ptr.prototype.get = function() {
		var self, x, x$1;
		self = this;
		if (self.domAttr.j.Text() === "undefined") {
			return (x = new StringEqualer.Ptr(""), new x.constructor.Struct(x));
		}
		return (x$1 = new StringEqualer.Ptr(self.domAttr.j.Text()), new x$1.constructor.Struct(x$1));
	};
	textAttr.prototype.get = function() { return this.go$val.get(); };
	textAttr.Ptr.prototype.set = function(e) {
		var self;
		self = this;
		self.domAttr.j.SetText(new Go$String(fmt.Sprintf("%v", new (go$sliceType(go$emptyInterface))([e]))));
	};
	textAttr.prototype.set = function(e) { return this.go$val.set(e); };
	NewHtmlAttrAttr = go$pkg.NewHtmlAttrAttr = function(j, a) {
		var result, _struct, _recv, e, _recv$1;
		result = new htmlAttrAttr.Ptr((go$ptrType(domAttr)).nil, a);
		result.domAttr = newDomAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), "attr:" + a, (_recv = result, function() { return _recv.get(); }), (_recv$1 = result, function(e) { return _recv$1.set(e); }));
		return result;
	};
	htmlAttrAttr.Ptr.prototype.get = function() {
		var self, x, x$1;
		self = this;
		if (self.domAttr.j.Attr(self.name) === "undefined") {
			return (x = new StringEqualer.Ptr(""), new x.constructor.Struct(x));
		}
		return (x$1 = new StringEqualer.Ptr(self.domAttr.j.Attr(self.name)), new x$1.constructor.Struct(x$1));
	};
	htmlAttrAttr.prototype.get = function() { return this.go$val.get(); };
	htmlAttrAttr.Ptr.prototype.set = function(e) {
		var self;
		self = this;
		self.domAttr.j.SetAttr(new (go$sliceType(go$emptyInterface))([new Go$String(self.name), new Go$String(fmt.Sprintf("%v", new (go$sliceType(go$emptyInterface))([e])))]));
	};
	htmlAttrAttr.prototype.set = function(e) { return this.go$val.set(e); };
	NewPropAttr = go$pkg.NewPropAttr = function(j, n) {
		var result, _struct, _recv, e, _recv$1;
		result = new propAttr.Ptr((go$ptrType(domAttr)).nil, n);
		result.domAttr = newDomAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), "prop:" + n, (_recv = result, function() { return _recv.get(); }), (_recv$1 = result, function(e) { return _recv$1.set(e); }));
		return result;
	};
	propAttr.Ptr.prototype.get = function() {
		var self, b, x;
		self = this;
		b = self.domAttr.j.Prop(self.name);
		return (x = new BoolEqualer.Ptr((b !== null && b.constructor === Go$Bool ? b.go$val : go$typeAssertionFailed(b, Go$Bool))), new x.constructor.Struct(x));
	};
	propAttr.prototype.get = function() { return this.go$val.get(); };
	propAttr.Ptr.prototype.set = function(e) {
		var self;
		self = this;
		self.domAttr.j.SetProp(new (go$sliceType(go$emptyInterface))([new Go$String(self.name), new Go$Bool((e !== null && e.constructor === BoolEqualer ? e.go$val : go$typeAssertionFailed(e, BoolEqualer)).B)]));
	};
	propAttr.prototype.set = function(e) { return this.go$val.set(e); };
	newAttrName = function(s) {
		return s;
	};
	newPropName = function(s) {
		return s;
	};
	NewCssExistenceAttr = go$pkg.NewCssExistenceAttr = function(j, clazz) {
		var result, _struct, _recv, e, _recv$1;
		result = new cssExistenceAttr.Ptr((go$ptrType(domAttr)).nil, clazz);
		result.domAttr = newDomAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), "cssclass:" + clazz.ClassName(), (_recv = result, function() { return _recv.get(); }), (_recv$1 = result, function(e) { return _recv$1.set(e); }));
		return result;
	};
	cssExistenceAttr.Ptr.prototype.get = function() {
		var self, x;
		self = this;
		return (x = new BoolEqualer.Ptr(self.domAttr.j.HasClass(self.clazz.ClassName())), new x.constructor.Struct(x));
	};
	cssExistenceAttr.prototype.get = function() { return this.go$val.get(); };
	cssExistenceAttr.Ptr.prototype.set = function(e) {
		var self;
		self = this;
		if ((e !== null && e.constructor === BoolEqualer ? e.go$val : go$typeAssertionFailed(e, BoolEqualer)).B) {
			self.domAttr.j.AddClass(new Go$String(self.clazz.ClassName()));
		} else {
			self.domAttr.j.RemoveClass(self.clazz.ClassName());
		}
	};
	cssExistenceAttr.prototype.set = function(e) { return this.go$val.set(e); };
	NewDisplayAttr = go$pkg.NewDisplayAttr = function(j) {
		var result, _struct, _recv, e, _recv$1;
		result = new styleAttr.Ptr((go$ptrType(domAttr)).nil, "display");
		result.domAttr = newDomAttr((_struct = j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), "style:display", (_recv = result, function() { return _recv.getDisplay(); }), (_recv$1 = result, function(e) { return _recv$1.setDisplay(e); }));
		return result;
	};
	styleAttr.Ptr.prototype.getDisplay = function() {
		var self, x, x$1;
		self = this;
		if (self.domAttr.j.Css("display") === "undefined") {
			return (x = new BoolEqualer.Ptr(true), new x.constructor.Struct(x));
		}
		return (x$1 = new BoolEqualer.Ptr(!(self.domAttr.j.Css("display") === "none")), new x$1.constructor.Struct(x$1));
	};
	styleAttr.prototype.getDisplay = function() { return this.go$val.getDisplay(); };
	styleAttr.Ptr.prototype.setDisplay = function(e) {
		var self;
		self = this;
		if ((e !== null && e.constructor === BoolEqualer ? e.go$val : go$typeAssertionFailed(e, BoolEqualer)).B) {
			self.domAttr.j.SetCss(new (go$sliceType(go$emptyInterface))([new Go$String("display"), new Go$String("")]));
		} else {
			self.domAttr.j.SetCss(new (go$sliceType(go$emptyInterface))([new Go$String("display"), new Go$String("none")]));
		}
	};
	styleAttr.prototype.setDisplay = function(e) { return this.go$val.setDisplay(e); };
	DrainEagerQueue = go$pkg.DrainEagerQueue = function() {
		var _ref, _i, a;
		_ref = eagerQueue;
		_i = 0;
		while (_i < _ref.length) {
			a = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			a.Demand();
			_i++;
		}
		eagerQueue = (go$sliceType((go$ptrType(AttributeImpl)))).nil;
	};
	newEdge = function(src, dest) {
		var e;
		e = new edgeImpl.Ptr(false, dest, src);
		src.addOut(e);
		dest.addIn(e);
		return e;
	};
	dropEdge = function(src, dest) {
		src.removeOut(dest.id());
		dest.removeIn(src.id());
		dest.markDirty();
	};
	edgeImpl.Ptr.prototype.marked = function() {
		var self;
		self = this;
		return !self.notMarked;
	};
	edgeImpl.prototype.marked = function() { return this.go$val.marked(); };
	edgeImpl.Ptr.prototype.mark = function() {
		var self;
		self = this;
		self.notMarked = false;
	};
	edgeImpl.prototype.mark = function() { return this.go$val.mark(); };
	edgeImpl.Ptr.prototype.src = function() {
		var self;
		self = this;
		return self.s;
	};
	edgeImpl.prototype.src = function() { return this.go$val.src(); };
	edgeImpl.Ptr.prototype.dest = function() {
		var self;
		self = this;
		return self.d;
	};
	edgeImpl.prototype.dest = function() { return this.go$val.dest(); };
	NewAttribute = go$pkg.NewAttribute = function(nt, v, s) {
		var result;
		if (nt > 2) {
			throw go$panic(new Go$String("unexpected node type"));
		}
		result = new AttributeImpl.Ptr(numNodes, 0, 0, false, (go$sliceType((go$ptrType(edgeImpl)))).nil, null, null, s, v, nt);
		numNodes = numNodes + 1 >> 0;
		return result;
	};
	AttributeImpl.Ptr.prototype.Attach = function(c) {
		var self, edges, _ref, _i, other, i;
		self = this;
		if (self.nType === 1) {
			throw go$panic(new Go$String("cant attach a constraint simple value!"));
		}
		if (!((self.inDegree() === 0))) {
			throw go$panic(new Go$String("constraint is already attached!"));
		}
		self.markDirty();
		self.constraint = c;
		edges = (go$sliceType((go$ptrType(edgeImpl)))).make(c.Inputs().length, 0, function() { return (go$ptrType(edgeImpl)).nil; });
		_ref = c.Inputs();
		_i = 0;
		while (_i < _ref.length) {
			other = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			(i < 0 || i >= edges.length) ? go$throwRuntimeError("index out of range") : edges.array[edges.offset + i] = newEdge((other !== null && node.implementedBy.indexOf(other.constructor) !== -1 ? other : go$typeAssertionFailed(other, node)), self);
			_i++;
		}
		if (self.nType === 2) {
			self.Demand();
		}
		DrainEagerQueue();
	};
	AttributeImpl.prototype.Attach = function(c) { return this.go$val.Attach(c); };
	AttributeImpl.Ptr.prototype.Detach = function() {
		var self, dead, _ref, _i, e;
		self = this;
		dead = new (go$sliceType((go$ptrType(edgeImpl))))([]);
		self.walkIncoming((function(e) {
			e.src().removeOut(self.id());
			dead = go$append(dead, e);
		}));
		_ref = dead;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			self.removeIn(e.src().id());
			_i++;
		}
		DrainEagerQueue();
	};
	AttributeImpl.prototype.Detach = function() { return this.go$val.Detach(); };
	AttributeImpl.Ptr.prototype.dropIthEdge = function(i) {
		var self, x, x$1, x$2, x$3, x$4;
		self = this;
		(x$2 = self.edge, (i < 0 || i >= x$2.length) ? go$throwRuntimeError("index out of range") : x$2.array[x$2.offset + i] = (x = self.edge, x$1 = self.edge.length - 1 >> 0, ((x$1 < 0 || x$1 >= x.length) ? go$throwRuntimeError("index out of range") : x.array[x.offset + x$1])));
		(x$3 = self.edge, x$4 = self.edge.length - 1 >> 0, (x$4 < 0 || x$4 >= x$3.length) ? go$throwRuntimeError("index out of range") : x$3.array[x$3.offset + x$4] = (go$ptrType(edgeImpl)).nil);
		self.edge = go$subslice(self.edge, 0, (self.edge.length - 1 >> 0));
	};
	AttributeImpl.prototype.dropIthEdge = function(i) { return this.go$val.dropIthEdge(i); };
	AttributeImpl.Ptr.prototype.removeIn = function(id) {
		var self, _ref, _i, e, i;
		self = this;
		_ref = self.edge;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if ((e.dest().id() === self.id()) && (e.src().id() === id)) {
				self.dropIthEdge(i);
				break;
			}
			_i++;
		}
	};
	AttributeImpl.prototype.removeIn = function(id) { return this.go$val.removeIn(id); };
	AttributeImpl.Ptr.prototype.removeOut = function(id) {
		var self, _ref, _i, e, i;
		self = this;
		_ref = self.edge;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			i = _i;
			if ((e.src().id() === self.id()) && (e.dest().id() === id)) {
				self.dropIthEdge(i);
				break;
			}
			_i++;
		}
	};
	AttributeImpl.prototype.removeOut = function(id) { return this.go$val.removeOut(id); };
	AttributeImpl.Ptr.prototype.id = function() {
		var self;
		self = this;
		return self.n;
	};
	AttributeImpl.prototype.id = function() { return this.go$val.id(); };
	AttributeImpl.Ptr.prototype.walkOutgoing = function(fn) {
		var self, _ref, _i, e;
		self = this;
		_ref = self.edge;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			if (!((e.src().id() === self.id()))) {
				_i++;
				continue;
			}
			fn(e);
			_i++;
		}
	};
	AttributeImpl.prototype.walkOutgoing = function(fn) { return this.go$val.walkOutgoing(fn); };
	AttributeImpl.Ptr.prototype.walkIncoming = function(fn) {
		var self, _ref, _i, e;
		self = this;
		_ref = self.edge;
		_i = 0;
		while (_i < _ref.length) {
			e = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			if (!((e.dest().id() === self.id()))) {
				_i++;
				continue;
			}
			fn(e);
			_i++;
		}
	};
	AttributeImpl.prototype.walkIncoming = function(fn) { return this.go$val.walkIncoming(fn); };
	AttributeImpl.Ptr.prototype.dirty = function() {
		var self;
		self = this;
		return !self.clean;
	};
	AttributeImpl.prototype.dirty = function() { return this.go$val.dirty(); };
	AttributeImpl.Ptr.prototype.markDirty = function() {
		var self;
		self = this;
		if (self.dirty()) {
			return;
		}
		self.clean = false;
		self.walkOutgoing((function(e) {
			e.dest().markDirty();
		}));
		if (self.nType === 2) {
			eagerQueue = go$append(eagerQueue, self);
		}
	};
	AttributeImpl.prototype.markDirty = function() { return this.go$val.markDirty(); };
	AttributeImpl.Ptr.prototype.assign = function(newval, wantSideEffect) {
		var self;
		self = this;
		if (wantSideEffect && !(self.sideEffectFn === go$throwNilPointerError)) {
			self.sideEffectFn(newval);
		}
		self.curr = newval;
		return newval;
	};
	AttributeImpl.prototype.assign = function(newval, wantSideEffect) { return this.go$val.assign(newval, wantSideEffect); };
	AttributeImpl.Ptr.prototype.addOut = function(e) {
		var self;
		self = this;
		if (!e.marked()) {
			throw go$panic(new Go$String("badly constructed edge, not marked (OUT)"));
		}
		self.edge = go$append(self.edge, e);
	};
	AttributeImpl.prototype.addOut = function(e) { return this.go$val.addOut(e); };
	AttributeImpl.Ptr.prototype.addIn = function(e) {
		var self;
		self = this;
		if (!e.marked()) {
			throw go$panic(new Go$String("badly constructed edge, not marked (IN)"));
		}
		self.edge = go$append(self.edge, e);
	};
	AttributeImpl.prototype.addIn = function(e) { return this.go$val.addIn(e); };
	AttributeImpl.Ptr.prototype.source = function() {
		var self;
		self = this;
		return self.inDegree() === 0;
	};
	AttributeImpl.prototype.source = function() { return this.go$val.source(); };
	AttributeImpl.Ptr.prototype.inDegree = function() {
		var self, inDegree;
		self = this;
		inDegree = 0;
		self.walkIncoming((function(e) {
			inDegree = inDegree + 1 >> 0;
		}));
		return inDegree;
	};
	AttributeImpl.prototype.inDegree = function() { return this.go$val.inDegree(); };
	AttributeImpl.Ptr.prototype.SetEqualer = function(i) {
		var self;
		self = this;
		if (!self.source()) {
			throw go$panic(new Go$String("can't set a value on something that is not a source!"));
		}
		if (!(go$interfaceIsEqual(self.curr, null)) && self.curr.Equal(i)) {
			return;
		}
		self.markDirty();
		self.walkOutgoing((function(e) {
			e.mark();
		}));
		self.assign(i, true);
		DrainEagerQueue();
	};
	AttributeImpl.prototype.SetEqualer = function(i) { return this.go$val.SetEqualer(i); };
	AttributeImpl.Ptr.prototype.Demand = function() {
		var self, params, pcount, anyMarks, newval;
		self = this;
		self.demandCount = self.demandCount + 1 >> 0;
		if (!self.dirty()) {
			return self.curr;
		}
		self.clean = true;
		params = (go$sliceType(Equaler)).make(self.inDegree(), 0, function() { return null; });
		pcount = 0;
		self.walkIncoming((function(e) {
			(pcount < 0 || pcount >= params.length) ? go$throwRuntimeError("index out of range") : params.array[params.offset + pcount] = e.src().Demand();
			pcount = pcount + 1 >> 0;
		}));
		anyMarks = false;
		self.walkIncoming((function(e) {
			if (e.marked()) {
				anyMarks = true;
			}
		}));
		if (!anyMarks && self.valueFn === go$throwNilPointerError) {
			return self.curr;
		}
		self.evalCount = self.evalCount + 1 >> 0;
		newval = null;
		if (!(self.valueFn === go$throwNilPointerError) && go$interfaceIsEqual(self.constraint, null)) {
			newval = self.valueFn();
		} else {
			newval = self.constraint.Fn(params);
		}
		if (!(go$interfaceIsEqual(self.curr, null)) && self.curr.Equal(newval)) {
			return self.curr;
		}
		self.walkOutgoing((function(e) {
			e.mark();
		}));
		return self.assign(newval, true);
	};
	AttributeImpl.prototype.Demand = function() { return this.go$val.Demand(); };
	EventName.prototype.String = function() {
		var self, _ref;
		self = this.go$val;
		_ref = self;
		if (_ref === 0) {
			return "blur";
		} else if (_ref === 1) {
			return "change";
		} else if (_ref === 2) {
			return "click";
		} else if (_ref === 3) {
			return "dblclick";
		} else if (_ref === 4) {
			return "focus";
		} else if (_ref === 5) {
			return "focusin";
		} else if (_ref === 6) {
			return "focusout";
		} else if (_ref === 7) {
			return "hover";
		} else if (_ref === 8) {
			return "keydown";
		} else if (_ref === 9) {
			return "keypress";
		} else if (_ref === 10) {
			return "keyup";
		}
		throw go$panic(new Go$String(fmt.Sprintf("unknown event name %v", new (go$sliceType(go$emptyInterface))([new EventName(self)]))));
	};
	go$ptrType(EventName).prototype.String = function() { return new EventName(this.go$get()).String(); };
	eventHandler.Ptr.prototype.register = function() {
		var self, event, _recv;
		self = this;
		self.j.On(new (go$sliceType(go$emptyInterface))([new Go$String((new EventName(self.name)).String()), new (go$funcType([jquery.Event], [], false))((_recv = self, function(event) { return _recv.handle(event); }))]));
	};
	eventHandler.prototype.register = function() { return this.go$val.register(); };
	eventHandler.Ptr.prototype.handle = function(event) {
		var self, _struct, _struct$1;
		self = this;
		self.fn((_struct = self.j, new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context)), (_struct$1 = event, new jquery.Event.Ptr(_struct$1.Object, _struct$1.KeyCode, _struct$1.Target, _struct$1.CurrentTarget, _struct$1.DelegateTarget, _struct$1.RelatedTarget, _struct$1.Data, _struct$1.Result, _struct$1.Which, _struct$1.Namespace, _struct$1.MetaKey, _struct$1.PageX, _struct$1.PageY, _struct$1.Type)));
		DrainEagerQueue();
	};
	eventHandler.prototype.handle = function(event) { return this.go$val.handle(event); };
	IntEqualer.Ptr.prototype.Equal = function(e) {
		var _struct, self;
		self = (_struct = this, new IntEqualer.Ptr(_struct.I));
		return self.I === (e !== null && e.constructor === IntEqualer ? e.go$val : go$typeAssertionFailed(e, IntEqualer)).I;
	};
	IntEqualer.prototype.Equal = function(e) { return this.go$val.Equal(e); };
	IntEqualer.Ptr.prototype.String = function() {
		var _struct, self;
		self = (_struct = this, new IntEqualer.Ptr(_struct.I));
		return fmt.Sprintf("%d", new (go$sliceType(go$emptyInterface))([new Go$Int(self.I)]));
	};
	IntEqualer.prototype.String = function() { return this.go$val.String(); };
	IntegerSimple.Ptr.prototype.Value = function() {
		var self, x;
		self = this;
		return (x = self.AttributeImpl.Demand(), (x !== null && x.constructor === IntEqualer ? x.go$val : go$typeAssertionFailed(x, IntEqualer))).I;
	};
	IntegerSimple.prototype.Value = function() { return this.go$val.Value(); };
	IntegerSimple.Ptr.prototype.Set = function(i) {
		var self, x;
		self = this;
		self.AttributeImpl.SetEqualer((x = new IntEqualer.Ptr(i), new x.constructor.Struct(x)));
	};
	IntegerSimple.prototype.Set = function(i) { return this.go$val.Set(i); };
	NewIntegerSimple = go$pkg.NewIntegerSimple = function(i) {
		var result;
		result = new IntegerSimple.Ptr(NewAttribute(0, go$throwNilPointerError, go$throwNilPointerError));
		result.Set(i);
		return result;
	};
	StringEqualer.Ptr.prototype.Equal = function(e) {
		var _struct, self;
		self = (_struct = this, new StringEqualer.Ptr(_struct.S));
		return self.S === (e !== null && e.constructor === StringEqualer ? e.go$val : go$typeAssertionFailed(e, StringEqualer)).S;
	};
	StringEqualer.prototype.Equal = function(e) { return this.go$val.Equal(e); };
	StringEqualer.Ptr.prototype.String = function() {
		var _struct, self;
		self = (_struct = this, new StringEqualer.Ptr(_struct.S));
		return fmt.Sprintf("%s", new (go$sliceType(go$emptyInterface))([new Go$String(self.S)]));
	};
	StringEqualer.prototype.String = function() { return this.go$val.String(); };
	StringSimple.Ptr.prototype.Value = function() {
		var self, x;
		self = this;
		return (x = self.AttributeImpl.Demand(), (x !== null && x.constructor === StringEqualer ? x.go$val : go$typeAssertionFailed(x, StringEqualer))).S;
	};
	StringSimple.prototype.Value = function() { return this.go$val.Value(); };
	StringSimple.Ptr.prototype.Set = function(s) {
		var self, x;
		self = this;
		self.AttributeImpl.SetEqualer((x = new StringEqualer.Ptr(s), new x.constructor.Struct(x)));
	};
	StringSimple.prototype.Set = function(s) { return this.go$val.Set(s); };
	NewStringSimple = go$pkg.NewStringSimple = function(s) {
		var result;
		result = new StringSimple.Ptr(NewAttribute(0, go$throwNilPointerError, go$throwNilPointerError));
		result.Set(s);
		return result;
	};
	eqConstraint.Ptr.prototype.Inputs = function() {
		var _struct, self;
		self = (_struct = this, new eqConstraint.Ptr(_struct.dep));
		return new (go$sliceType(Attribute))([self.dep]);
	};
	eqConstraint.prototype.Inputs = function() { return this.go$val.Inputs(); };
	eqConstraint.Ptr.prototype.Fn = function(in$1) {
		var _struct, self;
		self = (_struct = this, new eqConstraint.Ptr(_struct.dep));
		if (!((in$1.length === 1))) {
			throw go$panic(new Go$String("wrong number of inputs to equality constraint"));
		}
		return self.dep.Demand();
	};
	eqConstraint.prototype.Fn = function(in$1) { return this.go$val.Fn(in$1); };
	Equality = go$pkg.Equality = function(dest, src) {
		var x;
		dest.Attach((x = new eqConstraint.Ptr(src), new x.constructor.Struct(x)));
	};
	ModelImpl.Ptr.prototype.Id = function() {
		var self;
		self = this;
		return fmt.Sprintf("%s-%d", new (go$sliceType(go$emptyInterface))([new Go$String(self.t), new Go$Int(self.n)]));
	};
	ModelImpl.prototype.Id = function() { return this.go$val.Id(); };
	NewModelName = go$pkg.NewModelName = function(i) {
		var t, idx, result;
		t = fmt.Sprintf("%T", new (go$sliceType(go$emptyInterface))([i]));
		if (t.charCodeAt(0) === 42) {
			t = t.substring(1);
		}
		idx = strings.Index(t, ".");
		if (!((idx === -1))) {
			t = t.substring((idx + 1 >> 0));
		}
		result = new ModelImpl.Ptr(count, t);
		count = count + 1 >> 0;
		return result;
	};
	Event = go$pkg.Event = function(name, fn) {
		return (function(self) {
			self.event = go$append(self.event, new eventHandler.Ptr(name, new jquery.JQuery.Ptr(), fn));
			return self;
		});
	};
	Class = go$pkg.Class = function(cl) {
		return (function(self) {
			self.classes = go$append(self.classes, cl.ClassName());
			return self;
		});
	};
	ModelId = go$pkg.ModelId = function(m) {
		return (function(self) {
			self.id = HtmlIdFromModel(self.tag, m).Id();
			return self;
		});
	};
	addBuilder = function(b) {
		return (function(self) {
			self.builders = go$append(self.builders, b);
			return self;
		});
	};
	CssExistence = go$pkg.CssExistence = function(c, b) {
		return addBuilder(cssExistenceBuilder(c, b));
	};
	PropEqual = go$pkg.PropEqual = function(n, b) {
		return addBuilder(propBuilder(n, b, null));
	};
	HtmlAttrEqual = go$pkg.HtmlAttrEqual = function(h, attr) {
		return addBuilder(htmlAttrBuilder(h, attr, null));
	};
	TextEqual = go$pkg.TextEqual = function(attr) {
		return addBuilder(textAttrBuilder(attr, null));
	};
	HtmlIdFromModel = go$pkg.HtmlIdFromModel = function(tag$1, m) {
		var id;
		id = fmt.Sprintf("%s-%s", new (go$sliceType(go$emptyInterface))([new Go$String(strings.ToLower(tag$1)), new Go$String(m.Id())]));
		return new htmlIdImpl.Ptr(tag$1, id);
	};
	ViewImpl.Ptr.prototype.Build = function() {
		var p, id, classes, t, parsed, _struct, j, _ref, _i, child, x, _ref$1, _i$1, b, _struct$1, _ref$2, _i$2, h, _struct$2, _struct$3;
		p = this;
		id = "";
		classes = "";
		if (!(p.id === "")) {
			id = fmt.Sprintf(" id='%s'", new (go$sliceType(go$emptyInterface))([new Go$String(p.id)]));
		}
		if (!(p.classes === (go$sliceType(Go$String)).nil)) {
			classes = fmt.Sprintf(" class='%s'", new (go$sliceType(go$emptyInterface))([new Go$String(strings.Join(p.classes, " "))]));
		}
		t = "";
		if (p.text === "") {
			t = fmt.Sprintf("<%s%s%s/>", new (go$sliceType(go$emptyInterface))([new Go$String(p.tag), new Go$String(id), new Go$String(classes)]));
		} else {
			t = fmt.Sprintf("<%s%s%s>%s</%s>", new (go$sliceType(go$emptyInterface))([new Go$String(p.tag), new Go$String(id), new Go$String(classes), new Go$String(p.text), new Go$String(p.tag)]));
		}
		parsed = jquery.ParseHTML(t);
		j = (_struct = jquery.NewJQuery(new (go$sliceType(go$emptyInterface))([((0 < 0 || 0 >= parsed.length) ? go$throwRuntimeError("index out of range") : parsed.array[parsed.offset + 0])])), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		_ref = p.children;
		_i = 0;
		while (_i < _ref.length) {
			child = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
			j.Append(new (go$sliceType(go$emptyInterface))([(x = child.Build(), new x.constructor.Struct(x))]));
			_i++;
		}
		if (!(p.builders === (go$sliceType(builder)).nil)) {
			_ref$1 = p.builders;
			_i$1 = 0;
			while (_i$1 < _ref$1.length) {
				b = ((_i$1 < 0 || _i$1 >= _ref$1.length) ? go$throwRuntimeError("index out of range") : _ref$1.array[_ref$1.offset + _i$1]);
				b.build((_struct$1 = j, new jquery.JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context)));
				_i$1++;
			}
		}
		if (!(p.event === (go$sliceType((go$ptrType(eventHandler)))).nil)) {
			_ref$2 = p.event;
			_i$2 = 0;
			while (_i$2 < _ref$2.length) {
				h = ((_i$2 < 0 || _i$2 >= _ref$2.length) ? go$throwRuntimeError("index out of range") : _ref$2.array[_ref$2.offset + _i$2]);
				h.j = (_struct$2 = j, new jquery.JQuery.Ptr(_struct$2.o, _struct$2.Jquery, _struct$2.Selector, _struct$2.Length, _struct$2.Context));
				h.register();
				_i$2++;
			}
		}
		return (_struct$3 = j, new jquery.JQuery.Ptr(_struct$3.o, _struct$3.Jquery, _struct$3.Selector, _struct$3.Length, _struct$3.Context));
	};
	ViewImpl.prototype.Build = function() { return this.go$val.Build(); };
	DIV = go$pkg.DIV = function(obj) {
		return tag("div", obj);
	};
	INPUT = go$pkg.INPUT = function(obj) {
		return tag("input", obj);
	};
	LABEL = go$pkg.LABEL = function(obj) {
		return tag("label", obj);
	};
	LI = go$pkg.LI = function(obj) {
		return tag("li", obj);
	};
	BUTTON = go$pkg.BUTTON = function(obj) {
		return tag("button", obj);
	};
	tag = function(tagName, obj) {
		var p, i, _tuple, x, opt, ok, _tuple$1, x$1, v;
		p = new ViewImpl.Ptr(tagName, "", (go$sliceType(Go$String)).nil, (go$sliceType((go$ptrType(ViewImpl)))).nil, "", (go$sliceType(builder)).nil, (go$sliceType((go$ptrType(eventHandler)))).nil);
		i = 0;
		while (i < obj.length) {
			if (go$interfaceIsEqual(((i < 0 || i >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + i]), null)) {
				throw go$panic(new Go$String("cant use a nil value in view construction"));
			}
			_tuple = (x = ((i < 0 || i >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + i]), (x !== null && x.constructor === option ? [x.go$val, true] : [go$throwNilPointerError, false])); opt = _tuple[0]; ok = _tuple[1];
			if (ok) {
				opt(p);
				i = i + 1 >> 0;
				continue;
			}
			_tuple$1 = (x$1 = ((i < 0 || i >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + i]), (x$1 !== null && x$1.constructor === (go$ptrType(ViewImpl)) ? [x$1.go$val, true] : [(go$ptrType(ViewImpl)).nil, false])); v = _tuple$1[0]; ok = _tuple$1[1];
			if (ok) {
				p.children = go$append(p.children, v);
				i = i + 1 >> 0;
				continue;
			}
			throw go$panic(new Go$String(fmt.Sprintf("unable to understand type of parameter: %v (%T %d) to %s", new (go$sliceType(go$emptyInterface))([((i < 0 || i >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + i]), ((i < 0 || i >= obj.length) ? go$throwRuntimeError("index out of range") : obj.array[obj.offset + i]), new Go$Int(i), new Go$String(tagName)]))));
			i = i + 1 >> 0;
		}
		return p;
	};
	go$pkg.init = function() {
		BooleanAttribute.init([["Attach", "", (go$funcType([Constraint], [], false))], ["Demand", "", (go$funcType([], [Equaler], false))], ["Detach", "", (go$funcType([], [], false))], ["Set", "", (go$funcType([Go$Bool], [], false))], ["SetEqualer", "", (go$funcType([Equaler], [], false))], ["Value", "", (go$funcType([], [Go$Bool], false))]]);
		BooleanSimple.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		(go$ptrType(BooleanSimple)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Set", "", [Go$Bool], [], false, -1], ["SetEqualer", "", [Equaler], [], false, 0], ["Value", "", [], [Go$Bool], false, -1], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		BooleanSimple.init([["AttributeImpl", "", "", (go$ptrType(AttributeImpl)), ""]]);
		BoolEqualer.methods = [["Equal", "", [Equaler], [Go$Bool], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(BoolEqualer)).methods = [["Equal", "", [Equaler], [Go$Bool], false, -1], ["String", "", [], [Go$String], false, -1]];
		BoolEqualer.init([["B", "B", "", Go$Bool, ""]]);
		(go$ptrType(BooleanInverter)).methods = [["Fn", "", [(go$sliceType(Equaler))], [Equaler], false, -1], ["Inputs", "", [], [(go$sliceType(Attribute))], false, -1]];
		BooleanInverter.init([["dep", "dep", "todomvc/concorde", Attribute, ""]]);
		(go$ptrType(BoolEq)).methods = [["Fn", "", [(go$sliceType(Equaler))], [Equaler], false, -1], ["Inputs", "", [], [(go$sliceType(Attribute))], false, -1]];
		BoolEq.init([["dep", "dep", "todomvc/concorde", BooleanAttribute, ""]]);
		builder.init([["build", "todomvc/concorde", (go$funcType([jquery.JQuery], [], false))]]);
		getDomAttrFunc.init([jquery.JQuery], [DomAttribute], false);
		builderBase.init([["cons", "cons", "todomvc/concorde", Constraint, ""], ["get", "get", "todomvc/concorde", getDomAttrFunc, ""]]);
		(go$ptrType(anyBuilder)).methods = [["build", "todomvc/concorde", [jquery.JQuery], [], false, -1]];
		anyBuilder.init([["builderBase", "", "todomvc/concorde", builderBase, ""], ["a", "a", "todomvc/concorde", Attribute, ""]]);
		(go$ptrType(booleanBuilder)).methods = [["build", "todomvc/concorde", [jquery.JQuery], [], false, -1]];
		booleanBuilder.init([["builderBase", "", "todomvc/concorde", builderBase, ""], ["b", "b", "todomvc/concorde", BooleanAttribute, ""]]);
		(go$ptrType(Collection)).methods = [["Add", "", [Model], [], false, -1], ["All", "", [], [(go$sliceType(Model))], false, -1], ["AllFold", "", [Attribute, go$emptyInterface, Puller, FoldingIterator, Equaler], [Constraint], false, -1], ["EmptyAttribute", "", [], [BooleanAttribute], false, -1], ["LengthAttribute", "", [], [IntegerAttribute], false, -1], ["PopRaw", "", [], [Model], false, -1], ["PushRaw", "", [Model], [], false, -1], ["Remove", "", [Model], [], false, -1], ["empty", "todomvc/concorde", [], [Equaler], false, -1], ["length", "todomvc/concorde", [], [Equaler], false, -1]];
		Collection.init([["coll", "coll", "todomvc/concorde", (go$ptrType(AttributeImpl)), ""], ["join", "join", "todomvc/concorde", (go$sliceType(Joiner)), ""]]);
		Joiner.init([["Add", "", (go$funcType([Go$Int, Model], [], false))], ["Remove", "", (go$funcType([Go$Int, Model], [], false))]]);
		EqList.methods = [["Equal", "", [Equaler], [Go$Bool], false, -1]];
		(go$ptrType(EqList)).methods = [["Equal", "", [Equaler], [Go$Bool], false, -1]];
		EqList.init(Model);
		computedEmpty.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		(go$ptrType(computedEmpty)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Set", "", [Go$Bool], [], false, -1], ["SetEqualer", "", [Equaler], [], false, 0], ["Value", "", [], [Go$Bool], false, -1], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		computedEmpty.init([["AttributeImpl", "", "", (go$ptrType(AttributeImpl)), ""], ["coll", "coll", "todomvc/concorde", (go$ptrType(Collection)), ""]]);
		computedLength.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		(go$ptrType(computedLength)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Set", "", [Go$Int], [], false, -1], ["SetEqualer", "", [Equaler], [], false, 0], ["Value", "", [], [Go$Int], false, -1], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		computedLength.init([["AttributeImpl", "", "", (go$ptrType(AttributeImpl)), ""], ["coll", "coll", "todomvc/concorde", (go$ptrType(Collection)), ""]]);
		Puller.init([Model], [Attribute], false);
		FoldingIterator.init([go$emptyInterface, Equaler], [go$emptyInterface, Equaler], false);
		(go$ptrType(foldedConstraint)).methods = [["Add", "", [Go$Int, Model], [], false, -1], ["Fn", "", [(go$sliceType(Equaler))], [Equaler], false, -1], ["Inputs", "", [], [(go$sliceType(Attribute))], false, -1], ["Remove", "", [Go$Int, Model], [], false, -1]];
		foldedConstraint.init([["deps", "deps", "todomvc/concorde", (go$sliceType(Attribute)), ""], ["fn", "fn", "todomvc/concorde", FoldingIterator, ""], ["pull", "pull", "todomvc/concorde", Puller, ""], ["initial", "initial", "todomvc/concorde", go$emptyInterface, ""], ["targ", "targ", "todomvc/concorde", Attribute, ""], ["empty", "empty", "todomvc/concorde", Equaler, ""]]);
		Equaler.init([["Equal", "", (go$funcType([Equaler], [Go$Bool], false))]]);
		Attribute.init([["Attach", "", (go$funcType([Constraint], [], false))], ["Demand", "", (go$funcType([], [Equaler], false))], ["Detach", "", (go$funcType([], [], false))], ["SetEqualer", "", (go$funcType([Equaler], [], false))]]);
		Constraint.init([["Fn", "", (go$funcType([(go$sliceType(Equaler))], [Equaler], false))], ["Inputs", "", (go$funcType([], [(go$sliceType(Attribute))], false))]]);
		CssClass.init([["ClassName", "", (go$funcType([], [Go$String], false))]]);
		cssClassImpl.methods = [["ClassName", "", [], [Go$String], false, -1]];
		(go$ptrType(cssClassImpl)).methods = [["ClassName", "", [], [Go$String], false, -1]];
		cssClassImpl.init([["name", "name", "todomvc/concorde", Go$String, ""]]);
		(go$ptrType(htmlIdImpl)).methods = [["DisplayAttribute", "", [], [DomAttribute], false, -1], ["Event", "", [EventName, EventFunc], [], false, -1], ["Id", "", [], [Go$String], false, -1], ["Select", "", [], [jquery.JQuery], false, -1], ["StyleAttribute", "", [Go$String], [DomAttribute], false, -1], ["TagName", "", [], [Go$String], false, -1], ["TextAttribute", "", [], [DomAttribute], false, -1]];
		htmlIdImpl.init([["tag", "tag", "todomvc/concorde", Go$String, ""], ["id", "id", "todomvc/concorde", Go$String, ""]]);
		DomAttribute.init([["Attach", "", (go$funcType([Constraint], [], false))], ["Demand", "", (go$funcType([], [Equaler], false))], ["Detach", "", (go$funcType([], [], false))], ["Id", "", (go$funcType([], [Go$String], false))], ["SetEqualer", "", (go$funcType([Equaler], [], false))]]);
		(go$ptrType(domAttr)).methods = [["Attach", "", [Constraint], [], false, -1], ["Demand", "", [], [Equaler], false, -1], ["Detach", "", [], [], false, -1], ["Id", "", [], [Go$String], false, -1], ["SetEqualer", "", [Equaler], [], false, -1], ["getData", "todomvc/concorde", [], [Go$String], false, -1], ["removeData", "todomvc/concorde", [], [], false, -1], ["setData", "todomvc/concorde", [Go$String], [], false, -1], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, -1]];
		domAttr.init([["attr", "attr", "todomvc/concorde", (go$ptrType(AttributeImpl)), ""], ["j", "j", "todomvc/concorde", jquery.JQuery, ""], ["id", "id", "todomvc/concorde", Go$String, ""], ["set", "set", "todomvc/concorde", SideEffectFunc, ""], ["get", "get", "todomvc/concorde", ValueFunc, ""]]);
		styleAttr.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		(go$ptrType(styleAttr)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["get", "todomvc/concorde", [], [Equaler], false, -1], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["getDisplay", "todomvc/concorde", [], [Equaler], false, -1], ["removeData", "todomvc/concorde", [], [], false, 0], ["set", "todomvc/concorde", [Equaler], [], false, -1], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["setDisplay", "todomvc/concorde", [Equaler], [], false, -1], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		styleAttr.init([["domAttr", "", "todomvc/concorde", (go$ptrType(domAttr)), ""], ["name", "name", "todomvc/concorde", Go$String, ""]]);
		textAttr.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		(go$ptrType(textAttr)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["get", "todomvc/concorde", [], [Equaler], false, -1], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["set", "todomvc/concorde", [Equaler], [], false, -1], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		textAttr.init([["domAttr", "", "todomvc/concorde", (go$ptrType(domAttr)), ""]]);
		htmlAttrAttr.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		(go$ptrType(htmlAttrAttr)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["get", "todomvc/concorde", [], [Equaler], false, -1], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["set", "todomvc/concorde", [Equaler], [], false, -1], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		htmlAttrAttr.init([["domAttr", "", "todomvc/concorde", (go$ptrType(domAttr)), ""], ["name", "name", "todomvc/concorde", htmlAttrName, ""]]);
		propAttr.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		(go$ptrType(propAttr)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["get", "todomvc/concorde", [], [Equaler], false, -1], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["set", "todomvc/concorde", [Equaler], [], false, -1], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		propAttr.init([["domAttr", "", "todomvc/concorde", (go$ptrType(domAttr)), ""], ["name", "name", "todomvc/concorde", propName, ""]]);
		cssExistenceAttr.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		(go$ptrType(cssExistenceAttr)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Id", "", [], [Go$String], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["get", "todomvc/concorde", [], [Equaler], false, -1], ["getData", "todomvc/concorde", [], [Go$String], false, 0], ["removeData", "todomvc/concorde", [], [], false, 0], ["set", "todomvc/concorde", [Equaler], [], false, -1], ["setData", "todomvc/concorde", [Go$String], [], false, 0], ["verifyConstraint", "todomvc/concorde", [Go$Bool], [], false, 0]];
		cssExistenceAttr.init([["domAttr", "", "todomvc/concorde", (go$ptrType(domAttr)), ""], ["clazz", "clazz", "todomvc/concorde", CssClass, ""]]);
		(go$ptrType(edgeImpl)).methods = [["clear", "todomvc/concorde", [], [], false, -1], ["dest", "todomvc/concorde", [], [node], false, -1], ["mark", "todomvc/concorde", [], [], false, -1], ["marked", "todomvc/concorde", [], [Go$Bool], false, -1], ["src", "todomvc/concorde", [], [node], false, -1]];
		edgeImpl.init([["notMarked", "notMarked", "todomvc/concorde", Go$Bool, ""], ["d", "d", "todomvc/concorde", node, ""], ["s", "s", "todomvc/concorde", node, ""]]);
		node.init([["Attach", "", (go$funcType([Constraint], [], false))], ["Demand", "", (go$funcType([], [Equaler], false))], ["Detach", "", (go$funcType([], [], false))], ["SetEqualer", "", (go$funcType([Equaler], [], false))], ["addIn", "todomvc/concorde", (go$funcType([(go$ptrType(edgeImpl))], [], false))], ["addOut", "todomvc/concorde", (go$funcType([(go$ptrType(edgeImpl))], [], false))], ["dirty", "todomvc/concorde", (go$funcType([], [Go$Bool], false))], ["id", "todomvc/concorde", (go$funcType([], [Go$Int], false))], ["markDirty", "todomvc/concorde", (go$funcType([], [], false))], ["numEdges", "todomvc/concorde", (go$funcType([], [Go$Int], false))], ["removeIn", "todomvc/concorde", (go$funcType([Go$Int], [], false))], ["removeOut", "todomvc/concorde", (go$funcType([Go$Int], [], false))], ["source", "todomvc/concorde", (go$funcType([], [Go$Bool], false))], ["walkIncoming", "todomvc/concorde", (go$funcType([(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false))], ["walkOutgoing", "todomvc/concorde", (go$funcType([(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false))]]);
		(go$ptrType(AttributeImpl)).methods = [["Attach", "", [Constraint], [], false, -1], ["Demand", "", [], [Equaler], false, -1], ["Detach", "", [], [], false, -1], ["SetEqualer", "", [Equaler], [], false, -1], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, -1], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, -1], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, -1], ["dirty", "todomvc/concorde", [], [Go$Bool], false, -1], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, -1], ["id", "todomvc/concorde", [], [Go$Int], false, -1], ["inDegree", "todomvc/concorde", [], [Go$Int], false, -1], ["markDirty", "todomvc/concorde", [], [], false, -1], ["numEdges", "todomvc/concorde", [], [Go$Int], false, -1], ["removeIn", "todomvc/concorde", [Go$Int], [], false, -1], ["removeOut", "todomvc/concorde", [Go$Int], [], false, -1], ["source", "todomvc/concorde", [], [Go$Bool], false, -1], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, -1], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, -1]];
		AttributeImpl.init([["n", "n", "todomvc/concorde", Go$Int, ""], ["evalCount", "evalCount", "todomvc/concorde", Go$Int, ""], ["demandCount", "demandCount", "todomvc/concorde", Go$Int, ""], ["clean", "clean", "todomvc/concorde", Go$Bool, ""], ["edge", "edge", "todomvc/concorde", (go$sliceType((go$ptrType(edgeImpl)))), ""], ["constraint", "constraint", "todomvc/concorde", Constraint, ""], ["curr", "curr", "todomvc/concorde", Equaler, ""], ["sideEffectFn", "sideEffectFn", "todomvc/concorde", SideEffectFunc, ""], ["valueFn", "valueFn", "todomvc/concorde", ValueFunc, ""], ["nType", "nType", "todomvc/concorde", NodeType, ""]]);
		ValueFunc.init([], [Equaler], false);
		SideEffectFunc.init([Equaler], [], false);
		EventFunc.init([jquery.JQuery, jquery.Event], [], false);
		EventName.methods = [["String", "", [], [Go$String], false, -1]];
		(go$ptrType(EventName)).methods = [["String", "", [], [Go$String], false, -1]];
		(go$ptrType(eventHandler)).methods = [["handle", "todomvc/concorde", [jquery.Event], [], false, -1], ["register", "todomvc/concorde", [], [], false, -1]];
		eventHandler.init([["name", "name", "todomvc/concorde", EventName, ""], ["j", "j", "todomvc/concorde", jquery.JQuery, ""], ["fn", "fn", "todomvc/concorde", EventFunc, ""]]);
		IntegerAttribute.init([["Attach", "", (go$funcType([Constraint], [], false))], ["Demand", "", (go$funcType([], [Equaler], false))], ["Detach", "", (go$funcType([], [], false))], ["Set", "", (go$funcType([Go$Int], [], false))], ["SetEqualer", "", (go$funcType([Equaler], [], false))], ["Value", "", (go$funcType([], [Go$Int], false))]]);
		IntegerSimple.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		(go$ptrType(IntegerSimple)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Set", "", [Go$Int], [], false, -1], ["SetEqualer", "", [Equaler], [], false, 0], ["Value", "", [], [Go$Int], false, -1], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		IntegerSimple.init([["AttributeImpl", "", "", (go$ptrType(AttributeImpl)), ""]]);
		IntEqualer.methods = [["Equal", "", [Equaler], [Go$Bool], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(IntEqualer)).methods = [["Equal", "", [Equaler], [Go$Bool], false, -1], ["String", "", [], [Go$String], false, -1]];
		IntEqualer.init([["I", "I", "", Go$Int, ""]]);
		StringAttribute.init([["Attach", "", (go$funcType([Constraint], [], false))], ["Demand", "", (go$funcType([], [Equaler], false))], ["Detach", "", (go$funcType([], [], false))], ["Set", "", (go$funcType([Go$String], [], false))], ["SetEqualer", "", (go$funcType([Equaler], [], false))], ["Value", "", (go$funcType([], [Go$String], false))]]);
		StringSimple.methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["SetEqualer", "", [Equaler], [], false, 0], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		(go$ptrType(StringSimple)).methods = [["Attach", "", [Constraint], [], false, 0], ["Demand", "", [], [Equaler], false, 0], ["Detach", "", [], [], false, 0], ["Set", "", [Go$String], [], false, -1], ["SetEqualer", "", [Equaler], [], false, 0], ["Value", "", [], [Go$String], false, -1], ["addIn", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["addOut", "todomvc/concorde", [(go$ptrType(edgeImpl))], [], false, 0], ["assign", "todomvc/concorde", [Equaler, Go$Bool], [Equaler], false, 0], ["dirty", "todomvc/concorde", [], [Go$Bool], false, 0], ["dropIthEdge", "todomvc/concorde", [Go$Int], [], false, 0], ["id", "todomvc/concorde", [], [Go$Int], false, 0], ["inDegree", "todomvc/concorde", [], [Go$Int], false, 0], ["markDirty", "todomvc/concorde", [], [], false, 0], ["numEdges", "todomvc/concorde", [], [Go$Int], false, 0], ["removeIn", "todomvc/concorde", [Go$Int], [], false, 0], ["removeOut", "todomvc/concorde", [Go$Int], [], false, 0], ["source", "todomvc/concorde", [], [Go$Bool], false, 0], ["walkIncoming", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0], ["walkOutgoing", "todomvc/concorde", [(go$funcType([(go$ptrType(edgeImpl))], [], false))], [], false, 0]];
		StringSimple.init([["AttributeImpl", "", "", (go$ptrType(AttributeImpl)), ""]]);
		StringEqualer.methods = [["Equal", "", [Equaler], [Go$Bool], false, -1], ["String", "", [], [Go$String], false, -1]];
		(go$ptrType(StringEqualer)).methods = [["Equal", "", [Equaler], [Go$Bool], false, -1], ["String", "", [], [Go$String], false, -1]];
		StringEqualer.init([["S", "S", "", Go$String, ""]]);
		eqConstraint.methods = [["Fn", "", [(go$sliceType(Equaler))], [Equaler], false, -1], ["Inputs", "", [], [(go$sliceType(Attribute))], false, -1]];
		(go$ptrType(eqConstraint)).methods = [["Fn", "", [(go$sliceType(Equaler))], [Equaler], false, -1], ["Inputs", "", [], [(go$sliceType(Attribute))], false, -1]];
		eqConstraint.init([["dep", "dep", "todomvc/concorde", Attribute, ""]]);
		ModelName.init([["Id", "", (go$funcType([], [Go$String], false))]]);
		Model.init([["Equal", "", (go$funcType([Equaler], [Go$Bool], false))], ["Id", "", (go$funcType([], [Go$String], false))]]);
		(go$ptrType(ModelImpl)).methods = [["Id", "", [], [Go$String], false, -1]];
		ModelImpl.init([["n", "n", "todomvc/concorde", Go$Int, ""], ["t", "t", "todomvc/concorde", Go$String, ""]]);
		(go$ptrType(ViewImpl)).methods = [["Build", "", [], [jquery.JQuery], false, -1]];
		ViewImpl.init([["tag", "tag", "todomvc/concorde", Go$String, ""], ["id", "id", "todomvc/concorde", Go$String, ""], ["classes", "classes", "todomvc/concorde", (go$sliceType(Go$String)), ""], ["children", "children", "todomvc/concorde", (go$sliceType((go$ptrType(ViewImpl)))), ""], ["text", "text", "todomvc/concorde", Go$String, ""], ["builders", "builders", "todomvc/concorde", (go$sliceType(builder)), ""], ["event", "event", "todomvc/concorde", (go$sliceType((go$ptrType(eventHandler)))), ""]]);
		option.init([(go$ptrType(ViewImpl))], [(go$ptrType(ViewImpl))], false);
		go$pkg.REL = newAttrName("rel");
		go$pkg.LINK = newAttrName("link");
		go$pkg.TYPE = newAttrName("type");
		go$pkg.PLACEHOLDER = newAttrName("placeholder");
		go$pkg.HREF = newAttrName("href");
		go$pkg.SRC = newAttrName("src");
		go$pkg.VALUE = newAttrName("value");
		go$pkg.CHECKED = newPropName("checked");
		go$pkg.SELECTED = newPropName("selected");
		go$pkg.DISABLED = newPropName("disabled");
		eagerQueue = new (go$sliceType((go$ptrType(AttributeImpl))))([]);
		numNodes = 0;
		count = 0;
	}
	return go$pkg;
})();
go$packages["todomvc"] = (function() {
	var go$pkg = {}, jquery = go$packages["github.com/gopherjs/jquery"], strings = go$packages["strings"], concorde = go$packages["todomvc/concorde"], todo, todoApp, newTodo, newApp, main, sectionMain, footer, primaryInput, listContainer, pluralSpan, todoCount, clearCompleted, numCompleted, toggleAll, view, toggle, destroy, edit, completed, editing;
	todo = go$pkg.todo = go$newType(0, "Struct", "main.todo", "todo", "todomvc", function(modName_, name_, done_, editing_) {
		this.go$val = this;
		this.modName = modName_ !== undefined ? modName_ : null;
		this.name = name_ !== undefined ? name_ : null;
		this.done = done_ !== undefined ? done_ : null;
		this.editing = editing_ !== undefined ? editing_ : null;
	});
	todoApp = go$pkg.todoApp = go$newType(0, "Struct", "main.todoApp", "todoApp", "todomvc", function(todos_, numNotDone_, plural_, someDone_, numDone_, editing_) {
		this.go$val = this;
		this.todos = todos_ !== undefined ? todos_ : (go$ptrType(concorde.Collection)).nil;
		this.numNotDone = numNotDone_ !== undefined ? numNotDone_ : null;
		this.plural = plural_ !== undefined ? plural_ : null;
		this.someDone = someDone_ !== undefined ? someDone_ : null;
		this.numDone = numDone_ !== undefined ? numDone_ : null;
		this.editing = editing_ !== undefined ? editing_ : null;
	});
	todo.Ptr.prototype.Id = function() {
		var self;
		self = this;
		return self.modName.Id();
	};
	todo.prototype.Id = function() { return this.go$val.Id(); };
	todo.Ptr.prototype.Equal = function(e) {
		var self, other;
		self = this;
		if (go$interfaceIsEqual(e, null)) {
			return false;
		}
		other = (e !== null && e.constructor === (go$ptrType(todo)) ? e.go$val : go$typeAssertionFailed(e, (go$ptrType(todo))));
		return self.Id() === other.Id();
	};
	todo.prototype.Equal = function(e) { return this.go$val.Equal(e); };
	newTodo = function(raw) {
		var result;
		result = new todo.Ptr(null, concorde.NewStringSimple(raw), concorde.NewBooleanSimple(false), concorde.NewBooleanSimple(false));
		result.modName = concorde.NewModelName(result);
		return result;
	};
	todoApp.Ptr.prototype.Start = function() {
		var self, _struct, todoCountSelect, _struct$1;
		self = this;
		primaryInput.Event(1, (function(j, event) {
			if (!self.createTodo(j.Val())) {
				event.PreventDefault();
			}
		}));
		todoCountSelect = (_struct = todoCount.Select().Children(new Go$String("strong")), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		concorde.Equality(concorde.NewTextAttr((_struct$1 = todoCountSelect, new jquery.JQuery.Ptr(_struct$1.o, _struct$1.Jquery, _struct$1.Selector, _struct$1.Length, _struct$1.Context))), self.numNotDone);
		concorde.Equality(pluralSpan.TextAttribute(), self.plural);
		sectionMain.DisplayAttribute().Attach(concorde.NewBooleanInverter(self.todos.EmptyAttribute()));
		footer.DisplayAttribute().Attach(concorde.NewBooleanInverter(self.todos.EmptyAttribute()));
		concorde.Equality(clearCompleted.DisplayAttribute(), self.someDone);
		concorde.Equality(numCompleted.TextAttribute(), self.numDone);
		clearCompleted.Event(2, (function() {
			var all, dead, ct, _ref, _i, model, _ref$1, _i$1, d;
			all = self.todos.All();
			if (all.length === 0) {
				return;
			}
			dead = (go$sliceType(concorde.Model)).make(all.length, 0, function() { return null; });
			ct = 0;
			_ref = all;
			_i = 0;
			while (_i < _ref.length) {
				model = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
				if ((model !== null && model.constructor === (go$ptrType(todo)) ? model.go$val : go$typeAssertionFailed(model, (go$ptrType(todo)))).done.Value()) {
					(ct < 0 || ct >= dead.length) ? go$throwRuntimeError("index out of range") : dead.array[dead.offset + ct] = model;
					ct = ct + 1 >> 0;
				}
				_i++;
			}
			_ref$1 = dead;
			_i$1 = 0;
			while (_i$1 < _ref$1.length) {
				d = ((_i$1 < 0 || _i$1 >= _ref$1.length) ? go$throwRuntimeError("index out of range") : _ref$1.array[_ref$1.offset + _i$1]);
				self.todos.Remove(d);
				_i$1++;
			}
		}));
		toggleAll.Event(2, (function() {
			var desired, _ref, _i, m;
			desired = true;
			if (self.todos.LengthAttribute().Value() === self.numDone.Value()) {
				desired = false;
			}
			_ref = self.todos.All();
			_i = 0;
			while (_i < _ref.length) {
				m = ((_i < 0 || _i >= _ref.length) ? go$throwRuntimeError("index out of range") : _ref.array[_ref.offset + _i]);
				(m !== null && m.constructor === (go$ptrType(todo)) ? m.go$val : go$typeAssertionFailed(m, (go$ptrType(todo)))).done.Set(desired);
				_i++;
			}
		}));
		self.dependsOnAll();
	};
	todoApp.prototype.Start = function() { return this.go$val.Start(); };
	todoApp.Ptr.prototype.createTodo = function(v) {
		var self, result, todo$1;
		self = this;
		result = strings.Trim(v, " \n\t");
		if (result === "") {
			return false;
		}
		primaryInput.Select().SetVal(new Go$String(""));
		todo$1 = newTodo(result);
		self.todos.PushRaw(todo$1);
		return true;
	};
	todoApp.prototype.createTodo = function(v) { return this.go$val.createTodo(v); };
	newApp = function() {
		var result;
		result = new todoApp.Ptr((go$ptrType(concorde.Collection)).nil, null, null, null, null, null);
		result.todos = concorde.NewList(result);
		result.numNotDone = concorde.NewIntegerSimple(0);
		result.plural = concorde.NewStringSimple("");
		result.someDone = concorde.NewBooleanSimple(false);
		result.numDone = concorde.NewIntegerSimple(0);
		return result;
	};
	todoApp.Ptr.prototype.pullDone = function(m) {
		var self;
		self = this;
		return (m !== null && m.constructor === (go$ptrType(todo)) ? m.go$val : go$typeAssertionFailed(m, (go$ptrType(todo)))).done;
	};
	todoApp.prototype.pullDone = function(m) { return this.go$val.pullDone(m); };
	todoApp.Ptr.prototype.dependsOnAll = function() {
		var self, m, _recv, x, m$1, _recv$1, x$1, m$2, _recv$2, x$2, m$3, _recv$3, x$3;
		self = this;
		self.todos.AllFold(self.numNotDone, new Go$Int(0), (_recv = self, function(m) { return _recv.pullDone(m); }), (function(prev, curr) {
			var p, x;
			p = (prev !== null && prev.constructor === Go$Int ? prev.go$val : go$typeAssertionFailed(prev, Go$Int));
			if (!(curr !== null && curr.constructor === concorde.BoolEqualer ? curr.go$val : go$typeAssertionFailed(curr, concorde.BoolEqualer)).B) {
				p = p + 1 >> 0;
			}
			return [new Go$Int(p), (x = new concorde.IntEqualer.Ptr(p), new x.constructor.Struct(x))];
		}), (x = new concorde.IntEqualer.Ptr(0), new x.constructor.Struct(x)));
		self.todos.AllFold(self.plural, new Go$Int(0), (_recv$1 = self, function(m$1) { return _recv$1.pullDone(m$1); }), (function(prev, curr) {
			var p, s, x$1;
			p = (prev !== null && prev.constructor === Go$Int ? prev.go$val : go$typeAssertionFailed(prev, Go$Int));
			if (!(curr !== null && curr.constructor === concorde.BoolEqualer ? curr.go$val : go$typeAssertionFailed(curr, concorde.BoolEqualer)).B) {
				p = p + 1 >> 0;
			}
			s = "s";
			if (p === 1) {
				s = "";
			}
			return [new Go$Int(p), (x$1 = new concorde.StringEqualer.Ptr(s), new x$1.constructor.Struct(x$1))];
		}), (x$1 = new concorde.StringEqualer.Ptr("s"), new x$1.constructor.Struct(x$1)));
		self.todos.AllFold(self.someDone, new Go$Int(0), (_recv$2 = self, function(m$2) { return _recv$2.pullDone(m$2); }), (function(prev, curr) {
			var p, x$2;
			p = (prev !== null && prev.constructor === Go$Int ? prev.go$val : go$typeAssertionFailed(prev, Go$Int));
			if ((curr !== null && curr.constructor === concorde.BoolEqualer ? curr.go$val : go$typeAssertionFailed(curr, concorde.BoolEqualer)).B) {
				p = p + 1 >> 0;
			}
			return [new Go$Int(p), (x$2 = new concorde.BoolEqualer.Ptr(p > 0), new x$2.constructor.Struct(x$2))];
		}), (x$2 = new concorde.BoolEqualer.Ptr(false), new x$2.constructor.Struct(x$2)));
		self.todos.AllFold(self.numDone, new Go$Int(0), (_recv$3 = self, function(m$3) { return _recv$3.pullDone(m$3); }), (function(prev, curr) {
			var p, x$3;
			p = (prev !== null && prev.constructor === Go$Int ? prev.go$val : go$typeAssertionFailed(prev, Go$Int));
			if ((curr !== null && curr.constructor === concorde.BoolEqualer ? curr.go$val : go$typeAssertionFailed(curr, concorde.BoolEqualer)).B) {
				p = p + 1 >> 0;
			}
			return [new Go$Int(p), (x$3 = new concorde.IntEqualer.Ptr(p), new x$3.constructor.Struct(x$3))];
		}), (x$3 = new concorde.IntEqualer.Ptr(0), new x$3.constructor.Struct(x$3)));
	};
	todoApp.prototype.dependsOnAll = function() { return this.go$val.dependsOnAll(); };
	todoApp.Ptr.prototype.Add = function(length, newObj) {
		var self, model, _struct, tree;
		self = this;
		model = (newObj !== null && newObj.constructor === (go$ptrType(todo)) ? newObj.go$val : go$typeAssertionFailed(newObj, (go$ptrType(todo))));
		tree = (_struct = concorde.LI(new (go$sliceType(go$emptyInterface))([new concorde.option(concorde.ModelId(model)), new concorde.option(concorde.CssExistence(completed, model.done)), new concorde.option(concorde.CssExistence(editing, model.editing)), concorde.DIV(new (go$sliceType(go$emptyInterface))([new concorde.option(concorde.Class(view)), concorde.INPUT(new (go$sliceType(go$emptyInterface))([new concorde.option(concorde.Class(toggle)), new concorde.option(concorde.HtmlAttrEqual(concorde.TYPE, concorde.NewStringSimple("checkbox"))), new concorde.option(concorde.PropEqual(concorde.CHECKED, model.done)), new concorde.option(concorde.Event(1, (function(ignored, e) {
			model.done.Set(!model.done.Value());
		})))])), concorde.LABEL(new (go$sliceType(go$emptyInterface))([new concorde.option(concorde.TextEqual(model.name)), new concorde.option(concorde.Event(3, (function(ignored, ignored2) {
			var _struct, in$1;
			model.editing.Set(true);
			in$1 = (_struct = concorde.HtmlIdFromModel("INPUT", model).Select(), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
			in$1.SetVal(new Go$String(model.name.Value()));
			in$1.Focus();
			in$1.Select(new (go$sliceType(go$emptyInterface))([]));
		})))])), concorde.BUTTON(new (go$sliceType(go$emptyInterface))([new concorde.option(concorde.Class(destroy)), new concorde.option(concorde.Event(2, (function() {
			self.todos.Remove(model);
		})))]))])), concorde.INPUT(new (go$sliceType(go$emptyInterface))([new concorde.option(concorde.ModelId(model)), new concorde.option(concorde.Class(edit)), new concorde.option(concorde.HtmlAttrEqual(concorde.PLACEHOLDER, model.name)), new concorde.option(concorde.Event(8, (function(j, e) {
			var _ref, v;
			_ref = go$parseInt(e.Object.which) >> 0;
			if (_ref === 13) {
				v = strings.Trim(j.Val(), " \n\t");
				if (v === "") {
					self.todos.Remove(model);
				} else {
					model.name.Set(v);
				}
				j.Blur();
				model.editing.Set(false);
				primaryInput.Select().Focus();
			} else if (_ref === 27) {
				model.editing.Set(false);
				primaryInput.Select().Focus();
			}
		})))]))])).Build(), new jquery.JQuery.Ptr(_struct.o, _struct.Jquery, _struct.Selector, _struct.Length, _struct.Context));
		listContainer.Select().Append(new (go$sliceType(go$emptyInterface))([new tree.constructor.Struct(tree)]));
	};
	todoApp.prototype.Add = function(length, newObj) { return this.go$val.Add(length, newObj); };
	todoApp.Ptr.prototype.Remove = function(IGNORED, oldObj) {
		var self, model, finder;
		self = this;
		model = (oldObj !== null && oldObj.constructor === (go$ptrType(todo)) ? oldObj.go$val : go$typeAssertionFailed(oldObj, (go$ptrType(todo))));
		finder = concorde.HtmlIdFromModel("li", model);
		finder.Select().Remove(new (go$sliceType(go$emptyInterface))([]));
	};
	todoApp.prototype.Remove = function(IGNORED, oldObj) { return this.go$val.Remove(IGNORED, oldObj); };
	main = go$pkg.main = function() {
		concorde.Main(newApp());
	};
	go$pkg.init = function() {
		(go$ptrType(todo)).methods = [["Equal", "", [concorde.Equaler], [Go$Bool], false, -1], ["Id", "", [], [Go$String], false, -1]];
		todo.init([["modName", "modName", "todomvc", concorde.ModelName, ""], ["name", "name", "todomvc", concorde.StringAttribute, ""], ["done", "done", "todomvc", concorde.BooleanAttribute, ""], ["editing", "editing", "todomvc", concorde.BooleanAttribute, ""]]);
		(go$ptrType(todoApp)).methods = [["Add", "", [Go$Int, concorde.Model], [], false, -1], ["Remove", "", [Go$Int, concorde.Model], [], false, -1], ["Start", "", [], [], false, -1], ["createTodo", "todomvc", [Go$String], [Go$Bool], false, -1], ["dependsOnAll", "todomvc", [], [], false, -1], ["pullDone", "todomvc", [concorde.Model], [concorde.Attribute], false, -1]];
		todoApp.init([["todos", "todos", "todomvc", (go$ptrType(concorde.Collection)), ""], ["numNotDone", "numNotDone", "todomvc", concorde.IntegerAttribute, ""], ["plural", "plural", "todomvc", concorde.StringAttribute, ""], ["someDone", "someDone", "todomvc", concorde.BooleanAttribute, ""], ["numDone", "numDone", "todomvc", concorde.IntegerAttribute, ""], ["editing", "editing", "todomvc", concorde.BooleanAttribute, ""]]);
		sectionMain = concorde.NewHtmlId("section", "main");
		footer = concorde.NewHtmlId("footer", "footer");
		primaryInput = concorde.NewHtmlId("input", "new-todo");
		listContainer = concorde.NewHtmlId("ul", "todo-list");
		pluralSpan = concorde.NewHtmlId("span", "plural");
		todoCount = concorde.NewHtmlId("span", "todo-count");
		clearCompleted = concorde.NewHtmlId("button", "clear-completed");
		numCompleted = concorde.NewHtmlId("span", "num-completed");
		toggleAll = concorde.NewHtmlId("input", "toggle-all");
		view = concorde.NewCssClass("view");
		toggle = concorde.NewCssClass("toggle");
		destroy = concorde.NewCssClass("destroy");
		edit = concorde.NewCssClass("edit");
		completed = concorde.NewCssClass("completed");
		editing = concorde.NewCssClass("editing");
	}
	return go$pkg;
})();
go$error.implementedBy = [go$packages["errors"].errorString.Ptr, go$packages["github.com/gopherjs/gopherjs/js"].Error.Ptr, go$packages["os"].PathError.Ptr, go$packages["os"].SyscallError.Ptr, go$packages["reflect"].ValueError.Ptr, go$packages["runtime"].TypeAssertionError.Ptr, go$packages["runtime"].errorString, go$packages["syscall"].Errno, go$packages["time"].ParseError.Ptr, go$ptrType(go$packages["runtime"].errorString), go$ptrType(go$packages["syscall"].Errno)];
go$packages["github.com/gopherjs/gopherjs/js"].Object.implementedBy = [go$packages["github.com/gopherjs/gopherjs/js"].Error, go$packages["github.com/gopherjs/gopherjs/js"].Error.Ptr, go$packages["github.com/gopherjs/jquery"].Event, go$packages["github.com/gopherjs/jquery"].Event.Ptr];
go$packages["sync"].Locker.implementedBy = [go$packages["sync"].Mutex.Ptr, go$packages["sync"].RWMutex.Ptr, go$packages["sync"].rlocker.Ptr, go$packages["syscall"].mmapper.Ptr];
go$packages["io"].RuneReader.implementedBy = [go$packages["fmt"].ss.Ptr];
go$packages["os"].FileInfo.implementedBy = [go$packages["os"].fileStat.Ptr];
go$packages["reflect"].Type.implementedBy = [go$packages["reflect"].arrayType.Ptr, go$packages["reflect"].chanType.Ptr, go$packages["reflect"].funcType.Ptr, go$packages["reflect"].interfaceType.Ptr, go$packages["reflect"].mapType.Ptr, go$packages["reflect"].ptrType.Ptr, go$packages["reflect"].rtype.Ptr, go$packages["reflect"].sliceType.Ptr, go$packages["reflect"].structType.Ptr];
go$packages["fmt"].Formatter.implementedBy = [];
go$packages["fmt"].GoStringer.implementedBy = [];
go$packages["fmt"].State.implementedBy = [go$packages["fmt"].pp.Ptr];
go$packages["fmt"].Stringer.implementedBy = [go$packages["os"].FileMode, go$packages["reflect"].ChanDir, go$packages["reflect"].Kind, go$packages["reflect"].Value, go$packages["reflect"].Value.Ptr, go$packages["reflect"].arrayType.Ptr, go$packages["reflect"].chanType.Ptr, go$packages["reflect"].funcType.Ptr, go$packages["reflect"].interfaceType.Ptr, go$packages["reflect"].mapType.Ptr, go$packages["reflect"].ptrType.Ptr, go$packages["reflect"].rtype.Ptr, go$packages["reflect"].sliceType.Ptr, go$packages["reflect"].structType.Ptr, go$packages["strconv"].decimal.Ptr, go$packages["time"].Duration, go$packages["time"].Location.Ptr, go$packages["time"].Month, go$packages["time"].Time, go$packages["time"].Time.Ptr, go$packages["time"].Weekday, go$packages["todomvc/concorde"].BoolEqualer, go$packages["todomvc/concorde"].BoolEqualer.Ptr, go$packages["todomvc/concorde"].EventName, go$packages["todomvc/concorde"].IntEqualer, go$packages["todomvc/concorde"].IntEqualer.Ptr, go$packages["todomvc/concorde"].StringEqualer, go$packages["todomvc/concorde"].StringEqualer.Ptr, go$ptrType(go$packages["os"].FileMode), go$ptrType(go$packages["reflect"].ChanDir), go$ptrType(go$packages["reflect"].Kind), go$ptrType(go$packages["time"].Duration), go$ptrType(go$packages["time"].Month), go$ptrType(go$packages["time"].Weekday), go$ptrType(go$packages["todomvc/concorde"].EventName)];
go$packages["fmt"].runeUnreader.implementedBy = [go$packages["fmt"].ss.Ptr];
go$packages["todomvc/concorde"].Attribute.implementedBy = [go$packages["todomvc/concorde"].AttributeImpl.Ptr, go$packages["todomvc/concorde"].BooleanSimple, go$packages["todomvc/concorde"].BooleanSimple.Ptr, go$packages["todomvc/concorde"].IntegerSimple, go$packages["todomvc/concorde"].IntegerSimple.Ptr, go$packages["todomvc/concorde"].StringSimple, go$packages["todomvc/concorde"].StringSimple.Ptr, go$packages["todomvc/concorde"].computedEmpty, go$packages["todomvc/concorde"].computedEmpty.Ptr, go$packages["todomvc/concorde"].computedLength, go$packages["todomvc/concorde"].computedLength.Ptr, go$packages["todomvc/concorde"].cssExistenceAttr, go$packages["todomvc/concorde"].cssExistenceAttr.Ptr, go$packages["todomvc/concorde"].domAttr.Ptr, go$packages["todomvc/concorde"].htmlAttrAttr, go$packages["todomvc/concorde"].htmlAttrAttr.Ptr, go$packages["todomvc/concorde"].propAttr, go$packages["todomvc/concorde"].propAttr.Ptr, go$packages["todomvc/concorde"].styleAttr, go$packages["todomvc/concorde"].styleAttr.Ptr, go$packages["todomvc/concorde"].textAttr, go$packages["todomvc/concorde"].textAttr.Ptr];
go$packages["todomvc/concorde"].BooleanAttribute.implementedBy = [go$packages["todomvc/concorde"].BooleanSimple.Ptr, go$packages["todomvc/concorde"].computedEmpty.Ptr];
go$packages["todomvc/concorde"].Constraint.implementedBy = [go$packages["todomvc/concorde"].BoolEq.Ptr, go$packages["todomvc/concorde"].BooleanInverter.Ptr, go$packages["todomvc/concorde"].eqConstraint, go$packages["todomvc/concorde"].eqConstraint.Ptr, go$packages["todomvc/concorde"].foldedConstraint.Ptr];
go$packages["todomvc/concorde"].CssClass.implementedBy = [go$packages["todomvc/concorde"].cssClassImpl, go$packages["todomvc/concorde"].cssClassImpl.Ptr];
go$packages["todomvc/concorde"].DomAttribute.implementedBy = [go$packages["todomvc/concorde"].cssExistenceAttr, go$packages["todomvc/concorde"].cssExistenceAttr.Ptr, go$packages["todomvc/concorde"].domAttr.Ptr, go$packages["todomvc/concorde"].htmlAttrAttr, go$packages["todomvc/concorde"].htmlAttrAttr.Ptr, go$packages["todomvc/concorde"].propAttr, go$packages["todomvc/concorde"].propAttr.Ptr, go$packages["todomvc/concorde"].styleAttr, go$packages["todomvc/concorde"].styleAttr.Ptr, go$packages["todomvc/concorde"].textAttr, go$packages["todomvc/concorde"].textAttr.Ptr];
go$packages["todomvc/concorde"].Equaler.implementedBy = [go$packages["todomvc"].todo.Ptr, go$packages["todomvc/concorde"].BoolEqualer, go$packages["todomvc/concorde"].BoolEqualer.Ptr, go$packages["todomvc/concorde"].EqList, go$packages["todomvc/concorde"].IntEqualer, go$packages["todomvc/concorde"].IntEqualer.Ptr, go$packages["todomvc/concorde"].StringEqualer, go$packages["todomvc/concorde"].StringEqualer.Ptr, go$ptrType(go$packages["todomvc/concorde"].EqList)];
go$packages["todomvc/concorde"].IntegerAttribute.implementedBy = [go$packages["todomvc/concorde"].IntegerSimple.Ptr, go$packages["todomvc/concorde"].computedLength.Ptr];
go$packages["todomvc/concorde"].Joiner.implementedBy = [go$packages["todomvc"].todoApp.Ptr, go$packages["todomvc/concorde"].foldedConstraint.Ptr];
go$packages["todomvc/concorde"].Model.implementedBy = [go$packages["todomvc"].todo.Ptr];
go$packages["todomvc/concorde"].ModelName.implementedBy = [go$packages["todomvc"].todo.Ptr, go$packages["todomvc/concorde"].ModelImpl.Ptr, go$packages["todomvc/concorde"].cssExistenceAttr, go$packages["todomvc/concorde"].cssExistenceAttr.Ptr, go$packages["todomvc/concorde"].domAttr.Ptr, go$packages["todomvc/concorde"].htmlAttrAttr, go$packages["todomvc/concorde"].htmlAttrAttr.Ptr, go$packages["todomvc/concorde"].htmlIdImpl.Ptr, go$packages["todomvc/concorde"].propAttr, go$packages["todomvc/concorde"].propAttr.Ptr, go$packages["todomvc/concorde"].styleAttr, go$packages["todomvc/concorde"].styleAttr.Ptr, go$packages["todomvc/concorde"].textAttr, go$packages["todomvc/concorde"].textAttr.Ptr];
go$packages["todomvc/concorde"].StringAttribute.implementedBy = [go$packages["todomvc/concorde"].StringSimple.Ptr];
go$packages["todomvc/concorde"].builder.implementedBy = [go$packages["todomvc/concorde"].anyBuilder.Ptr, go$packages["todomvc/concorde"].booleanBuilder.Ptr];
go$packages["todomvc/concorde"].node.implementedBy = [go$packages["todomvc/concorde"].AttributeImpl.Ptr, go$packages["todomvc/concorde"].BooleanSimple, go$packages["todomvc/concorde"].BooleanSimple.Ptr, go$packages["todomvc/concorde"].IntegerSimple, go$packages["todomvc/concorde"].IntegerSimple.Ptr, go$packages["todomvc/concorde"].StringSimple, go$packages["todomvc/concorde"].StringSimple.Ptr, go$packages["todomvc/concorde"].computedEmpty, go$packages["todomvc/concorde"].computedEmpty.Ptr, go$packages["todomvc/concorde"].computedLength, go$packages["todomvc/concorde"].computedLength.Ptr];
go$packages["github.com/gopherjs/gopherjs/js"].init();
go$packages["runtime"].init();
go$packages["github.com/gopherjs/jquery"].init();
go$packages["errors"].init();
go$packages["sync/atomic"].init();
go$packages["sync"].init();
go$packages["io"].init();
go$packages["unicode"].init();
go$packages["unicode/utf8"].init();
go$packages["strings"].init();
go$packages["math"].init();
go$packages["syscall"].init();
go$packages["time"].init();
go$packages["os"].init();
go$packages["strconv"].init();
go$packages["reflect"].init();
go$packages["fmt"].init();
go$packages["todomvc/concorde"].init();
go$packages["todomvc"].init();
go$packages["todomvc"].main();

})();
//# sourceMappingURL=todomvc.js.map
