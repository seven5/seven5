#/bin/bash
for i in auth.go env.go page.go
do
mockgen --source=$i --package=auth --destination=mock_$i
done
