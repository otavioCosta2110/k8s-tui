#!/bin/bash
files=(
    "internal/app/ui/models/cronjobs.go"
    "internal/app/ui/models/daemonsets.go"
    "internal/app/ui/models/ingresses.go"
    "internal/app/ui/models/jobs.go"
    "internal/app/ui/models/nodes.go"
    "internal/app/ui/models/replicasets.go"
    "internal/app/ui/models/secrets.go"
    "internal/app/ui/models/serviceaccounts.go"
    "internal/app/ui/models/statefulsets.go"
)

for file in "${files[@]}"; do
    # Fix fetchFunc
    sed -i 's/fetchFunc := func() (\[\]table\.Row, error) {\s*return .*dataToRows(), nil\s*}/fetchFunc := func() ([]table.Row, error) {\n\t\tif err := \1.fetchData(); err != nil {\n\t\t\treturn nil, err\n\t\t}\n\t\treturn \1.dataToRows(), nil\n\t}/' "$file"
    
    # Fix table creation if not already fixed
    sed -i 's/tableModel := ui\.NewTable([^,]*, [^,]*, [^,]*\.dataToRows(),/tableModel := ui.NewTable(\1, \2, []table.Row{},/' "$file"
done
