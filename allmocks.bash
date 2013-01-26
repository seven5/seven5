#/bin/bash
for i in cookie.go session.go auth.go env.go page.go
do
mockgen --source=$i --package=seven5 --destination=mock_$i
done

 