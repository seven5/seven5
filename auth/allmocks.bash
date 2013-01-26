#/bin/bash
for i in auth.go env.go page.go
do
mockgen --source=$i --package=seven5 --destination=mock_$i
done
