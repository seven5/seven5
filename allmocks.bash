#/bin/bash
for i in cookie.go session.go
do
mockgen --source=$i --package=seven5 --destination=mock_$i
done

 