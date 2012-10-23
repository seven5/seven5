
This directory contains dart _source_ code that will be converted into go _source_ code before
distribution of the library.  This code usually is converted with the use of

    seven5tool embedfile filename

The result of embedding `foo.dart` in the code is `foo.dart.go` in the seven5 library.