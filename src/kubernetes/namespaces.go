package kubernetes

type Namespace struct {
	Name     string
	Children []Resource
}

// maybe change children to not initialize here, but be fetched in a function
func newNamespace(name string, children []Resource) Namespace {
	return Namespace{
    Name: name,
    Children: children,
  }
}

func FetchNamespaces(){
  println("bunda")
}
