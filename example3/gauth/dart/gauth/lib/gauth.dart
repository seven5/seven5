library gauth;
import 'package:gauth/seven5.dart';

class GauthUser {
		int Id;
	 
		String Name;
	 
		String GoogleId;
	 
		String Email;
	 
		String Pic;
		


	static String resourceURL = "/gauthuser/";

	static void Index(Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Index(resourceURL, ()=>new List<GauthUser>(), ()=>new GauthUser(), successFunc, errorFunc, headers, requestParameters);
	}

	static void Delete(int id, Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Delete(id, resourceURL, new GauthUser(), successFunc, errorFunc, headers, requestParameters);
	}

	void Put(Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Put(JSON.stringify(this), Id, resourceURL, new GauthUser(), successFunc, errorFunc, headers, requestParameters);
	}

	static void Post(dynamic example, Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Post(JSON.stringify(example), resourceURL, new GauthUser(), successFunc, errorFunc, headers, requestParameters);
	}

	void Find(int id, Function successFunc, [Function errorFunc, Map headers, Map requestParameters]) {
		Seven5Support.Find(id, resourceURL, new GauthUser(), successFunc, errorFunc, headers, requestParameters);
	}
	
	//convenience constructor
	GauthUser.fromJson(Map json) {
		copyFromJson(json);
	}
	
	//nothing to do in default constructor
	GauthUser();
	
	//this is the "magic" that changes from untyped Json to typed object
	GauthUser copyFromJson(Map json) {
		
	

		
			this.Id = json["Id"];
			
		

		
			this.Name = json["Name"];
			
		

		
			this.GoogleId = json["GoogleId"];
			
		

		
			this.Email = json["Email"];
			
		

		
			this.Pic = json["Pic"];
			
		 

		return this;
	}
	
	
Map toMapForJson() {
	Map result = new Map();
	
		result['Id']=Id;
	
		result['Name']=Name;
	
		result['GoogleId']=GoogleId;
	
		result['Email']=Email;
	
		result['Pic']=Pic;
	 
	return result;
}

	
	//this converts the object to a map so JSON serialization will like it
	toJson() {
		try {
			return this.toMapForJson();
		} catch (e) {
			print("something went wrong during JSON encoding: ${e}");
		}
	}
}

