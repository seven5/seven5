package seven5

const classdecl_tmpl = `
{{define "FIELD_DECL"}}
	{{range .Struct}} 
		{{.Dart}} {{.Name}};
	{{end}}	
{{end}}

{{define "COPY_JSON_FIELDS"}}
	{{range .Struct}}

		{{if .StructName}}
			this.{{.Name}} = new {{.StructName}}.fromJson(json["{{.Name}}"]);
		{{else}}
			this.{{.Name}} = json["{{.Name}}"];
			{{end}}{{/* if */}}
		{{end}} {{/* range */}}
{{end}} {{/* define */}}

{{define "EMIT_JSON_FIELDS"}}
Map toMapForJson() {
	Map result = new Map();
	{{range .Struct}}
		result['{{.Name}}']={{.Name}};
	{{end}} {{/* range */}}
	return result;
}
{{end}} {{/* define */}}

class {{.Name}} {
	{{template "FIELD_DECL" .}}

	static String resourceURL = "{{.RestPrefix}}{{tolower .Name}}/";

	static void Index(Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Index(resourceURL, ()=>new List<{{.Name}}>(), ()=>new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}

	static void Delete(int id, Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Delete(id, resourceURL, new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}

	void Put(Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Put(JSON.stringify(this), Id, resourceURL, new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}

	static void Post(dynamic example, Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Post(JSON.stringify(example), resourceURL, new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}

	void Find(int id, Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Find(id, resourceURL, new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}
	
	//convenience constructor
	{{.Name}}.fromJson(Map json) {
		copyFromJson(json);
	}
	
	//nothing to do in default constructor
	{{.Name}}();
	
	//this is the "magic" that changes from untyped Json to typed object
	{{.Name}} copyFromJson(Map json) {
		{{template "COPY_JSON_FIELDS" .}}
		return this;
	}
	
	{{template "EMIT_JSON_FIELDS" .}}
	
	//this converts the object to a map so JSON serialization will like it
	toJson() {
		try {
			return this.toMapForJson();
		} catch (e) {
			print("something went wrong during JSON encoding: ${e}");
		}
	}
}

{{define "SUPPORT_STRUCT_TMPL"}}
	class {{.StructName}} {
		{{template "FIELD_DECL" .}}

		//convenience constructor
		{{.StructName}}.fromJson(Map json) {
			copyFromJson(json);
		}
	
		//nothing to do in default constructor
		{{.StructName}}();
	
		//this is the "magic" that changes from untyped Json to typed object
		copyFromJson(Map json) {
			{{template "COPY_JSON_FIELDS" .}}
			return this;
		}
		{{template "EMIT_JSON_FIELDS" .}}
		
		//this converts the object to a map so JSON serialization will like it
		toJson() {
			return this.toMapForJson();
		}
		
	}
{{end}} {{/*define*/}}`
