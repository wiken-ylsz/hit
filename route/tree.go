package route

import (
	"net/http"
	"strings"
)

func newTree() *tree {
	return &tree{
		children: make(map[string]*tree),
		node:     new(node),
	}
}

type tree struct {
	children map[string]*tree // 子节点
	node     *node            // 包含的节点
}

func (t *tree) Add(path string, method string, h Handles) {
	var paths []string
	for _, val := range strings.Split(strings.ToUpper(path), "/") {
		if val == "" {
			continue
		}
		paths = append(paths, val)
	}
	method = strings.ToUpper(method)
	t.add(paths, HTTPMethod(method), h)
}

func (t *tree) add(path []string, method HTTPMethod, h Handles) {
	if len(path) == 0 {
		if t.node == nil {
			t.node = new(node)
		}
		t.node.add(method, h)
		return
	}
	if t.children[path[0]] == nil {
		t.children[path[0]] = newTree()
	}
	t.children[path[0]].add(path[1:], method, h)
}

func (t *tree) Find(path, method string) (h Handles) {
	var paths []string
	for _, val := range strings.Split(strings.ToUpper(path), "/") {
		if val == "" {
			continue
		}
		paths = append(paths, val)
	}
	method = strings.ToUpper(method)
	return t.find(paths, method)
}

func (t *tree) find(path []string, method HTTPMethod) (h Handles) {
	if len(path) == 0 {
		return t.node.get(method)
	}
	child := t.children[path[0]]
	if child == nil {
		return
	}

	return child.find(path[1:], method)
}

// HTTPMethod http请求方法包装
type HTTPMethod = string

// 常用的HTTP请求方法
const (
	MethodPost HTTPMethod = http.MethodPost
	MethodGet             = http.MethodGet
)

type node struct {
	// 实现REST请求
	postHandels Handles // Post请求处理方法
	getHandels  Handles // Get请求处理方法
}

func (n *node) get(method HTTPMethod) (h Handles) {
	switch method {
	case MethodPost:
		h = n.postHandels
	case MethodGet:
		h = n.getHandels
	default:
		return nil
	}
	return h
}

func (n *node) add(method HTTPMethod, h Handles) {
	switch method {
	case MethodPost:
		n.postHandels = append(n.postHandels, h...)
	case MethodGet:
		n.getHandels = append(n.getHandels, h...)
	default:
	}
}
