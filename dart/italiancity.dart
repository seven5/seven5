class LatLng {
	
	double Latitude;
	double Longitude; 
	
	LatLng.fromJson(Map json) {
		copyFromJson(json);
	}
	
	LatLng copyFromJson(Map json) {
		this.Latitude = json["Latitude"];
		this.Longitude = json["Longitude"];
		return this;
	}
}
class ItalianCity {
	int Id = Seven5Support.NOT_FETCHED;
	String Name;
	int Population;
	String Province;
	LatLng Location;
	
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
		this.LatLng = LatLng.fromJson(json["Location"]);
		return this;
	}
}
