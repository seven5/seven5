#import("dart:json");
#import("dart:io");

main() {
    HTTPClientConnection conn = HTTPClient.get("localhost","8887","/api/v1/schema");
}

