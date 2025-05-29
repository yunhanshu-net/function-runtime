package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	//repo, err := git.PlainInit("./repo", false)
	//if err != nil {
	//	panic(err)
	//}
	//create, err := os.CreateNode("./repo/test.txt")
	//if err != nil {
	//	panic(err)
	//}
	//defer create.Close()
	//create.WriteString("hello world\n")
	//worktree, err := repo.Worktree()
	//if err != nil {
	//	panic(err)
	//}
	//add, err := worktree.Add(".")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Add hash", add)
	//
	//commitMsg := "Initial commit"
	//h, err := worktree.Commit(commitMsg, &git.CommitOptions{
	//	Author: &object.Signature{
	//		Name:  "beiluo",
	//		Email: "beiluo@email.com",
	//		When:  time.Now(),
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Commit:", h)

	repo, err := git.PlainOpen("./repo")
	if err != nil {
		panic(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		panic(err)
	}
	err = worktree.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: plumbing.NewHash("HEAD^")})
	if err != nil {
		panic(err)
	}

	//add, err := worktree.Add(".")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Add hash", add)
	//
	//commitMsg := "Initial commit"
	//h, err := worktree.Commit(commitMsg, &git.CommitOptions{
	//	Author: &object.Signature{
	//		Name:  "beiluo",
	//		Email: "beiluo@email.com",
	//		When:  time.Now(),
	//	},
	//})
	//fmt.Println(h)

	// 获取 HEAD 引用
	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}

	// 创建提交遍历器
	commitIter, _ := repo.Log(&git.LogOptions{From: ref.Hash()})

	// 遍历提交记录
	_ = commitIter.ForEach(func(c *object.Commit) error {
		fmt.Printf("Commit: %s\nAuthor: %s\nDate: %s\nMessage: %s\n\n",
			c.Hash.String(),
			c.Author.String(),
			c.Author.When.Format("2006-01-02 15:04"),
			c.Message)
		return nil
	})

}
