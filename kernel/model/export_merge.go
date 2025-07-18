// SiYuan - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/siyuan-note/siyuan/kernel/filesys"
	"github.com/siyuan-note/siyuan/kernel/treenode"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func mergeSubDocs(rootTree *parse.Tree) (ret *parse.Tree, err error) {
	ret = rootTree
	rootBlock := &Block{Box: rootTree.Box, ID: rootTree.ID, Path: rootTree.Path, HPath: rootTree.HPath}
	if err = buildBlockChildren(rootBlock); err != nil {
		return
	}

	insertPoint := rootTree.Root.LastChild

	// 跳过空段落插入点，向上寻找非空段落
	for ; nil != insertPoint && ast.NodeParagraph == insertPoint.Type; insertPoint = insertPoint.Previous {
		if nil != insertPoint.FirstChild {
			break
		}
	}

	// 导出空文档 Word 和 PDF 时合并子文档失败 https://github.com/siyuan-note/siyuan/issues/7429
	if nil == insertPoint {
		// 如果找不到非空段落，则使用第一个段落作为插入点
		insertPoint = rootTree.Root.FirstChild
		if nil == insertPoint {
			// 如果文档为空，则创建一个空段落作为插入点
			insertPoint = treenode.NewParagraph("")
			rootTree.Root.AppendChild(insertPoint)
		}
	}

	for {
		i := 0
		if err = walkBlock(insertPoint, rootBlock, i); err != nil {
			return
		}
		if nil == rootBlock.Children {
			break
		}
	}

	if ast.NodeParagraph == insertPoint.Type && nil == insertPoint.FirstChild {
		// 删除空段落
		// Ignore the last empty paragraph block when exporting merged sub-documents https://github.com/siyuan-note/siyuan/issues/15028
		insertPoint.Unlink()
	}
	return
}

func walkBlock(insertPoint *ast.Node, block *Block, level int) (err error) {
	level++
	for i := len(block.Children) - 1; i >= 0; i-- {
		c := block.Children[i]
		if err = walkBlock(insertPoint, c, level); err != nil {
			return
		}

		nodes, loadErr := loadTreeNodes(c.Box, c.Path, level)
		if nil != loadErr {
			return
		}

		lastIndex := len(nodes) - 1
		for j := lastIndex; -1 < j; j-- {
			node := nodes[j]
			if j == lastIndex && ast.NodeParagraph == node.Type && nil == node.FirstChild {
				// 跳过最后一个空段落块
				// Ignore the last empty paragraph block when exporting merged sub-documents https://github.com/siyuan-note/siyuan/issues/15028
				continue
			}
			insertPoint.InsertAfter(node)
		}
	}
	block.Children = nil
	return
}

func loadTreeNodes(box string, p string, level int) (ret []*ast.Node, err error) {
	luteEngine := NewLute()
	tree, err := filesys.LoadTree(box, p, luteEngine)
	if err != nil {
		return
	}

	hLevel := level
	if 6 < level {
		hLevel = 6
	}

	heading := &ast.Node{ID: tree.Root.ID, Type: ast.NodeHeading, HeadingLevel: hLevel}
	heading.AppendChild(&ast.Node{Type: ast.NodeText, Tokens: []byte(tree.Root.IALAttr("title"))})
	tree.Root.PrependChild(heading)
	for c := tree.Root.FirstChild; nil != c; c = c.Next {
		ret = append(ret, c)
	}
	return
}

func buildBlockChildren(block *Block) (err error) {
	files, _, err := ListDocTree(block.Box, block.Path, util.SortModeUnassigned, false, false, Conf.FileTree.MaxListCount)
	if err != nil {
		return
	}

	for _, f := range files {
		childBlock := &Block{Box: block.Box, ID: f.ID, Path: f.Path}
		block.Children = append(block.Children, childBlock)
	}

	for _, c := range block.Children {
		if err = buildBlockChildren(c); err != nil {
			return
		}
	}
	return
}
