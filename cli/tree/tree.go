package tree

import (
	"fmt"
	"sort"
	"strings"
)

const (
	GroupIcon    = "🏷️"  // Icon for top-level group folders
	FolderIcon   = "📁"
	PasswordIcon = "🔑"
	BranchLast   = "└── "
	BranchMid    = "├── "
	Pipe         = "│   "
	Space        = "    "
)

type Node struct {
	Name     string
	IsFolder bool
	Children map[string]*Node
}

func NewNode(name string, isFolder bool) *Node {
	return &Node{
		Name:     name,
		IsFolder: isFolder,
		Children: make(map[string]*Node),
	}
}

func BuildTree(paths []string, rootName string) *Node {
	root := NewNode(rootName, true)
	
	for _, path := range paths {
		parts := strings.Split(path, "/")
		current := root
		
		for i, part := range parts {
			if part == "" {
				continue
			}
			
			isFolder := i < len(parts)-1
			
			if current.Children[part] == nil {
				current.Children[part] = NewNode(part, isFolder)
			}
			
			current = current.Children[part]
		}
	}
	
	return root
}

func (n *Node) Print(prefix string, isLast bool) {
	icon := PasswordIcon
	if n.IsFolder {
		icon = FolderIcon
	}
	
	var connector string
	if isLast {
		connector = BranchLast
	} else {
		connector = BranchMid
	}
	
	fmt.Printf("%s%s%s %s\n", prefix, connector, icon, n.Name)
	
	children := make([]*Node, 0, len(n.Children))
	names := make([]string, 0, len(n.Children))
	
	for name := range n.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	
	for _, name := range names {
		children = append(children, n.Children[name])
	}
	
	for i, child := range children {
		isChildLast := i == len(children)-1
		
		var childPrefix string
		if isLast {
			childPrefix = prefix + Space
		} else {
			childPrefix = prefix + Pipe
		}
		
		child.Print(childPrefix, isChildLast)
	}
}

func PrintTree(paths []string, rootName string) {
	if len(paths) == 0 {
		fmt.Printf("%s %s\n", GroupIcon, rootName)
		return
	}
	
	root := BuildTree(paths, rootName)
	fmt.Printf("%s %s\n", GroupIcon, root.Name)
	
	children := make([]*Node, 0, len(root.Children))
	names := make([]string, 0, len(root.Children))
	
	for name := range root.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	
	for _, name := range names {
		children = append(children, root.Children[name])
	}
	
	for i, child := range children {
		isLast := i == len(children)-1
		child.Print("", isLast)
	}
}