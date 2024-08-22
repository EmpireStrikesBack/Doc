package main

import "fmt"

func table() {
	a := [...]float64{12, 90, 2, 4}
	sum := float64(0)
	for i, v := range a {
		fmt.Printf("%d ème élément de a est %.2f\n", i, v)
		sum += v
	}
	fmt.Println("\n somme de tous les éléments de a est", sum)
}

func table1(a [3][2]string) {
	for _, v1 := range a {
		for _, v2 := range v1 {
			fmt.Printf("%s", v2)
		}
		fmt.Printf("\n")
	}
}

func main() {
	a := [3][2]string{
		{"jaune ", "rouge"},
		{"vert ", "mauve"},
		{"noir ", "blanc"},
	}
	table1(a)
	var b [3][2]string
	b[0][0] = "prada"
	b[0][1] = "lacoste"
	b[1][0] = "louboutin"
	b[1][1] = "gucci"
	b[2][0] = "dolce & gabanna"
	b[2][1] = "louis vuiton"
}
