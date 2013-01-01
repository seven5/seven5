library gauth;
import '/generated/dart';
import 'dart:html';
import 'dart:json';
import 'package:web_ui/watcher.dart' as watchers;

class App {
	//me and meta are defined to be types that are machine-generated from the go structs defined
	//on the server side.
	static GauthUser me = null;
	static GauthUserMetadata meta = null;

	//when a staff member needs to edit the users, the user list is kept here
	static List<GauthUser> all = null;
	//when a staff member edits the details of a particular user
	static GauthUser detail = null;

	//LoadSelf is the primary entry point for the App.  It is called from each page to ping 
	//the server and get information about the current user.  Nothing about the current user
	//is cached/stored, except the session id that is kept in a cookie.   The self=true parameter
	//here is to avoid loading the full list of users in the case where we are staff.
	static void LoadSelf(bool returnToHome) {
		GauthUser.Index(App.LoggedIn, (HttpRequest req) => App.NotLoggedIn(req, returnToHome), 
			null, {"self": "true"});
	}
	//NotLoggedIn is called if the attempt to read the data from the server in LoadSelf failed.
	//This indicates either we have no session or our session has expired (e.g. the server the
	//was rebooted and forgot the list of sessions).  The second parameter is true in cases where
	//we want this failure to be logged in to also result in taking the user's browser back to
	//the home page.
	static void NotLoggedIn(HttpRequest req, bool returnToHome) {
		if (req.status==403) {
			App.ChangeSelf(null);
			if (returnToHome) {
				document.window.location.href = "/out/home.html";
			}
		} else {
			print("Unexpected error trying to check for logged in: ${req.status}, ${req.responseText}");
		}
	}
	//LoggedIn is called with a list of size one resulting from a successful attempt to check for
	//being logged in (from LoadSelf()). The user's local data is updated, resulting in a UI
	//change, and then we attempt to see if this is user is also staff.
	static void LoggedIn(List<GauthUser> meList, HttpRequest req) {
		App.ChangeSelf(meList[0]);
		App.CheckForStaff();
	}
	//ChangeSelf is called on success or failure of login to update the local information about
	//the current information.  Because we are running in the background as a result of a networking
	//call, we must indicate the update to the watchers.
	static void ChangeSelf(GauthUser user) {
		App.me = user;
		watchers.dispatch();
	}
	//CheckForStaff is called to try to read the user metadata resource.  This will fail except
	//when the user is a staff member.  
	static void CheckForStaff() {
		GauthUserMetadata.Index(App.IsStaff, App.NotStaff, null, null);
	}
	//IsStaff is called when we have successfully read the metadata about the users, indicating that
	//we are staff.
	static void IsStaff(List<GauthUserMetadata> l, HttpRequest req) {
		App.ChangeMetadata(l[0]);
		App.GetAllUsers();
	}
	//NotStaff is called when we have failed to read the metadata about all users. This indicates
	//that the current user is not staff.
	static void NotStaff(HttpRequest req) {
		if (req.status==403) {
			App.ChangeMetadata(null);
		} else {
			print("Unexpected error trying to check for staff: ${req.status}, ${req.responseText}");
		}
	}
	//ChangeMetadata is called when we have a new copy of the user metadata to be displayed.  Only
	//staff members will see any UI resulting from this object.  This notifies the watchers that
	//the change has been made because we are updating this in the background as a result of a
	//networking call.
	static void ChangeMetadata(GauthUserMetadata m) {
		App.meta = m;
		watchers.dispatch();
	}
	//UpdateSelf is called when the user presses the "Submit" button on the form that changes
	//the user's data or when the user's data is updated by a staff member. The parameter will
	//have its data already set by the web component, so all that is necessary is call Put() 
	//to push it to the server.
	static void UpdateUser(GauthUser user) {
		user.Put(App.UpdateOk, App.UpdateError);
	}
	//CancelEdit is called when the user presses the "Cancel" button on the form to edit their
	//own data.  Because this immediately loads another page, the data (the changes) in 
	//data.me will be lost.
	static void CancelEdit() {
		document.window.location.href = "/out/home.html";
	}
	//UpdateOk is called when we have received confirmation from the server that our PUT has succeeded.
	//Any changes that were successfully made are in "changed".  Note that we are quickly shifting 
	//to home.html because that will refresh the information in App.me.
	static void UpdateOk(GauthUser changed, HttpRequest req) {
		App.ChangeSelf(changed);
		document.window.location.href = "/out/home.html";
	}
	//UpdateError is called when our PUT to update our values has failed.  This is not expected so
	//we just print out an error message.
	static void UpdateError(HttpRequest req) {
		print("Unexpected error trying to update userdata: ${req.status}, ${req.responseText}");
	}
	//Get all users is called if it appears that we are a staf user.  This retrieves the full list
	//of users, so it could be slow in the limit.  Note that auth is checked at the server side,
	//so there is no advantage to calling this if you are not staff.
	static void GetAllUsers() {
		GauthUser.Index(App.AllUsers, null, null, null);
	}
	//AllUsers is called if we succeed in called the method GET on the user resource for an index
	//of all users.  We call dispatch because we are 
	static void AllUsers(List<GauthUser> l, HttpRequest ignored) {
		App.ChangeAllUsers(l);
	}
	//ChangeAllUsers is called to indicate the value of all users is different than before.  This
	//needs to call the watchers because we are running in response to a network call (background).
	static void ChangeAllUsers(List<GauthUser> a) {
		App.all = a;
		watchers.dispatch();
	}
	//SelectUser is passed the google id of a user that we want to edit the details of.   
	static void SelectUser(GauthUser u) {
		App.detail = u;
	}
}