class ItalianCity {
	int Id = Seven5Support.NOT_FETCHED;
	String Name;
	int Population;
	String Province;
	
	static String findURL = "/italiancity/";
	static String indexURL = "/italiancities/";

	static List<ItalianCity> Index(successFunc, [errorFunc, headers, requestParameters]) {
		Seven5Support.Index(indexURL, ()=>new List<ItalianCity>(), 
			()=>new ItalianCity(), successFunc, errorFunc, headers, requestParameters);
	}

	void Find(int Id, successFunc, [errorFunc, headers, requestParameters]) {
		Seven5Support.Find(Id, findURL, new ItalianCity(), successFunc, errorFunc, headers,
			requestParameters);
	}
	
	//convenience constructor
	ItalianCity.fromJson(Map json) {
		copyFromJson(json);
	}
	
	//nothing to do in default constructor
	ItalianCity();
	
	//this is the "magic" that changes from untyped Json to typed object
	ItalianCity copyFromJson(Map json) {
		this.Id = json["Id"];
		this.Name = json["Name"];
		this.Population = json["Population"];
		this.Province = json["Province"];
		return this;
	}
}
