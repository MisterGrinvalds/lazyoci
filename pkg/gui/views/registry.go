package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// RegistryView displays the registry tree navigation
type RegistryView struct {
	TreeView *tview.TreeView
	registry *registry.Client
	onSelect func(repo string)
	root     *tview.TreeNode
}

// NewRegistryView creates a new registry tree view
func NewRegistryView(reg *registry.Client, onSelect func(repo string)) *RegistryView {
	rv := &RegistryView{
		registry: reg,
		onSelect: onSelect,
	}

	rv.root = tview.NewTreeNode("Registries").SetColor(tview.Styles.SecondaryTextColor)
	rv.TreeView = tview.NewTreeView().
		SetRoot(rv.root).
		SetCurrentNode(rv.root)

	rv.TreeView.SetBorder(true).SetTitle(" [1] Registries ")
	rv.TreeView.SetSelectedFunc(rv.handleSelect)

	rv.loadRegistries()

	return rv
}

func (rv *RegistryView) loadRegistries() {
	registries := rv.registry.GetRegistries()

	for _, reg := range registries {
		node := tview.NewTreeNode(reg.Name).
			SetReference(registryRef{url: reg.URL, name: reg.Name}).
			SetSelectable(true).
			SetExpanded(false)
		node.SetColor(tview.Styles.PrimaryTextColor)
		rv.root.AddChild(node)
	}
}

func (rv *RegistryView) handleSelect(node *tview.TreeNode) {
	ref := node.GetReference()
	if ref == nil {
		return
	}

	switch r := ref.(type) {
	case registryRef:
		if !node.IsExpanded() {
			rv.loadNamespaces(node, r.url)
		}
		node.SetExpanded(!node.IsExpanded())

	case namespaceRef:
		if !node.IsExpanded() {
			rv.loadRepositories(node, r.registry, r.namespace)
		}
		node.SetExpanded(!node.IsExpanded())

	case repositoryRef:
		if rv.onSelect != nil {
			rv.onSelect(r.fullPath)
		}
	}
}

func (rv *RegistryView) loadNamespaces(parent *tview.TreeNode, registryURL string) {
	parent.ClearChildren()

	namespaces, err := rv.registry.ListNamespaces(registryURL)
	if err != nil {
		infoNode := tview.NewTreeNode("[gray]Use / to search public registries[-]").
			SetSelectable(false)
		parent.AddChild(infoNode)
		return
	}

	if len(namespaces) == 0 {
		infoNode := tview.NewTreeNode("[gray]No namespaces found[-]").
			SetSelectable(false)
		parent.AddChild(infoNode)
		return
	}

	for _, ns := range namespaces {
		node := tview.NewTreeNode(ns).
			SetReference(namespaceRef{registry: registryURL, namespace: ns}).
			SetSelectable(true)
		parent.AddChild(node)
	}
}

func (rv *RegistryView) loadRepositories(parent *tview.TreeNode, registryURL, namespace string) {
	parent.ClearChildren()

	repos, err := rv.registry.ListRepositories(registryURL, namespace)
	if err != nil {
		errorNode := tview.NewTreeNode("[red]Error: " + err.Error() + "[-]").SetSelectable(false)
		parent.AddChild(errorNode)
		return
	}

	if len(repos) == 0 {
		infoNode := tview.NewTreeNode("[gray]No repositories found[-]").
			SetSelectable(false)
		parent.AddChild(infoNode)
		return
	}

	for _, repo := range repos {
		fullPath := registryURL + "/" + namespace + "/" + repo
		node := tview.NewTreeNode(repo).
			SetReference(repositoryRef{fullPath: fullPath}).
			SetSelectable(true)
		parent.AddChild(node)
	}
}

// AddRepository adds a repository node directly (for search results)
func (rv *RegistryView) AddRepository(registryURL, repoPath string) {
	for _, child := range rv.root.GetChildren() {
		ref := child.GetReference()
		if r, ok := ref.(registryRef); ok && r.url == registryURL {
			fullPath := registryURL + "/" + repoPath
			node := tview.NewTreeNode(repoPath).
				SetReference(repositoryRef{fullPath: fullPath}).
				SetSelectable(true).
				SetColor(tcell.ColorYellow)
			child.AddChild(node)
			child.SetExpanded(true)
			rv.TreeView.SetCurrentNode(node)
			return
		}
	}
}

// GetSelectedRegistry returns the URL of the currently selected or parent registry
func (rv *RegistryView) GetSelectedRegistry() string {
	node := rv.TreeView.GetCurrentNode()
	if node == nil {
		return ""
	}

	// Walk up the tree to find the registry
	for node != nil {
		ref := node.GetReference()
		switch r := ref.(type) {
		case registryRef:
			return r.url
		case namespaceRef:
			return r.registry
		case repositoryRef:
			// Parse registry from fullPath (e.g., "docker.io/library/nginx")
			parts := splitFirst(r.fullPath, "/")
			if len(parts) > 0 {
				return parts[0]
			}
		}
		node = findParent(rv.root, node)
	}
	return ""
}

func findParent(root, target *tview.TreeNode) *tview.TreeNode {
	if root == nil {
		return nil
	}
	for _, child := range root.GetChildren() {
		if child == target {
			return root
		}
		if found := findParent(child, target); found != nil {
			return found
		}
	}
	return nil
}

func splitFirst(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// Reference types for tree nodes
type registryRef struct {
	url  string
	name string
}

type namespaceRef struct {
	registry  string
	namespace string
}

type repositoryRef struct {
	fullPath string
}
