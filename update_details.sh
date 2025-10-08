#!/bin/bash

# List of files to update
files=(
    "internal/app/ui/models/cronjob_details.go"
    "internal/app/ui/models/daemonset_details.go"
    "internal/app/ui/models/ingress_details.go"
    "internal/app/ui/models/job_details.go"
    "internal/app/ui/models/statefulset_details.go"
)

for file in "${files[@]}"; do
    echo "Updating $file..."
    
    # Get the model name from the file
    model_name=$(grep "type.*DetailsModel struct" "$file" | sed 's/type \([a-zA-Z]*\)DetailsModel struct.*/\1/')
    echo "Model name: $model_name"
    
    # Add yamlViewer field to struct
    sed -i 's/loading   bool/loading   bool\n\tyamlViewer \*components.YAMLViewer/' "$file"
    
    # Add the loaded message type
    sed -i "s/type ${model_name}DetailsModel struct {/type ${model_name}DetailsLoadedMsg struct {\n\tcontent string\n\terr     error\n}\n\ntype ${model_name}DetailsModel struct {/" "$file"
    
    # Update New function to initialize yamlViewer
    title_prefix=$(echo "$model_name" | sed 's/.*/\u&/')
    sed -i "s/func New${title_prefix}Details(k k8s.Client, namespace, ${model_name}Name string) \*${model_name}DetailsModel {/func New${title_prefix}Details(k k8s.Client, namespace, ${model_name}Name string) \*${model_name}DetailsModel {\n\treturn \&${model_name}DetailsModel{\n\t\t${model_name}:   k8s.New${title_prefix}(${model_name}Name, namespace, k),\n\t\tk8sClient:  \&k,\n\t\tyamlViewer: components.NewYAMLViewer(\"${title_prefix}: \"+${model_name}Name, \"Loading...\"),\n\t\tloading:    true,\n\t\terr:        nil,\n\t}\n}/" "$file"
    
    # Remove the old New function body
    sed -i '/return \&'"${model_name}"'DetailsModel{/,/}/d' "$file"
    
    # Update InitComponent
    sed -i 's/func (.*) InitComponent(k \*k8s.Client) (tea.Model, error) {/,+2d; /return components.NewYAMLViewer/,+2d' "$file"
    sed -i 's/func (.*) InitComponent(k \*k8s.Client) (tea.Model, error) {/\t\1.k8sClient = k\n\treturn \1, nil\n}/' "$file"
    
    # Add fetch method
    sed -i 's/return \1, nil\n}/return \1, nil\n}\n\nfunc (\1) fetch'"${title_prefix}"'Details() tea.Cmd {\n\treturn func() tea.Msg {\n\t\tpm := plugins.GetGlobalPluginManager()\n\t\tapi := pm.GetAPI()\n\t\tapi.SetClient(*\1.k8sClient)\n\t\tdesc, err := api.Describe'"${title_prefix}"'(\1.'"${model_name}"'.Namespace, \1.'"${model_name}"'.Name)\n\t\treturn '"${model_name}"'DetailsLoadedMsg{content: desc, err: err}\n\t}\n}/' "$file"
    
    # Add Init method
    sed -i 's/}\n}\n\nfunc (\1) fetch/\n}\n\nfunc (\1) Init() tea.Cmd {\n\treturn tea.Batch(\1.yamlViewer.Init(), \1.fetch'"${title_prefix}"'Details())\n}\n\nfunc (\1) fetch/' "$file"
    
    # Add Update method
    sed -i 's/}\n}\n\nfunc (\1) Init/\n}\n\nfunc (\1) Update(msg tea.Msg) (tea.Model, tea.Cmd) {\n\tswitch msg := msg.(type) {\n\tcase '"${model_name}"'DetailsLoadedMsg:\n\t\t\1.loading = false\n\t\tif msg.err != nil {\n\t\t\t\1.err = msg.err\n\t\t\t\1.yamlViewer.SetContent("Error loading '"${model_name}"' details: " + msg.err.Error())\n\t\t} else {\n\t\t\t\1.yamlViewer.SetContent(msg.content)\n\t\t}\n\t\treturn \1, nil\n\tdefault:\n\t\tvar cmd tea.Cmd\n\t\tupdatedViewer, cmd := \1.yamlViewer.Update(msg)\n\t\t\1.yamlViewer = updatedViewer.(*components.YAMLViewer)\n\t\treturn \1, cmd\n\t}\n}\n\nfunc (\1) View() string {\n\treturn \1.yamlViewer.View()\n}\n\nfunc (\1) Init/' "$file"
    
done
