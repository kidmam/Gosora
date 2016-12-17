package main
import "log"
import "fmt"
import "strings"
import "strconv"
import "reflect"
import "path/filepath"
import "io/ioutil"
import "text/template/parse"

type VarItem struct
{
	Name string
	Destination string
	Type string
}

type VarItemReflect struct
{
	Name string
	Destination string
	Value reflect.Value
}

type CTemplateSet struct
{
	tlist map[string]*parse.Tree
	dir string
	funcMap map[string]interface{}
	importMap map[string]string
	varList map[string]VarItem
	localVars map[string]map[string]VarItemReflect
	stats map[string]int
	pVarList string
	pVarPosition int
	//tempVars map[string]string
	doImports bool
	expectsInt interface{}
}

func (c *CTemplateSet) compile_template(name string, dir string, expects string, expectsInt interface{}, varList map[string]VarItem) (out string) {
	c.dir = dir
	c.doImports = true
	c.funcMap = make(map[string]interface{})
	c.funcMap["and"] = "&&"
	c.funcMap["not"] = "!"
	c.funcMap["or"] = "||"
	c.funcMap["eq"] = true
	c.funcMap["ge"] = true
	c.funcMap["gt"] = true
	c.funcMap["le"] = true
	c.funcMap["lt"] = true
	c.funcMap["ne"] = true
	c.importMap = make(map[string]string)
	c.importMap["io"] = "io"
	c.importMap["strconv"] = "strconv"
	c.varList = varList
	//c.pVarList = ""
	//c.pVarPosition = 0
	c.stats = make(map[string]int)
	c.expectsInt = expectsInt
	holdreflect := reflect.ValueOf(expectsInt)
	
	res, err := ioutil.ReadFile(dir + name)
	if err != nil {
		log.Fatal(err)
	}
	content := string(res)
	
	tree := parse.New(name, c.funcMap)
	var treeSet map[string]*parse.Tree = make(map[string]*parse.Tree)
	tree, err = tree.Parse(content,"{{","}}", treeSet, c.funcMap)
	if err != nil {
		log.Fatal(err)
	}
	
	if debug {
		fmt.Println(name)
	}
	
	out = ""
	fname := strings.TrimSuffix(name, filepath.Ext(name))
	c.tlist = make(map[string]*parse.Tree)
	c.tlist[fname] = tree
	varholder := "tmpl_" + fname + "_vars"
	
	if debug {
		fmt.Println(c.tlist)
	}
	
	c.localVars = make(map[string]map[string]VarItemReflect)
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".",varholder,holdreflect}
	
	subtree := c.tlist[fname]
	if debug {
		fmt.Println(subtree.Root)
	}
	
	for _, node := range subtree.Root.Nodes {
		if debug {
			fmt.Println("Node: " + node.String())
		}
		out += c.compile_switch(varholder, holdreflect, fname, node)
	}
	
	var importList string
	if c.doImports {
		for _, item := range c.importMap {
			importList += "import \"" + item + "\"\n"
		}
	}
	
	var varString string
	for _, varItem := range c.varList {
		varString += "var " + varItem.Name + " " + varItem.Type + " = " + varItem.Destination + "\n"
	}
	
	out = "package main\n" + importList + c.pVarList + "\nfunc init() {\nctemplates[\"" + fname + "\"] = template_" + fname + "\n}\n\nfunc template_" + fname + "(tmpl_" + fname + "_vars " + expects + ", w io.Writer) {\n" + varString + out + "}\n"
	
	out = strings.Replace(out,`))
w.Write([]byte(`," + ",-1)
	out = strings.Replace(out,"` + `","",-1)
	
	for index, count := range c.stats {
		fmt.Println(index + ": " + strconv.Itoa(count))
	}
	
	if debug {
		fmt.Println("Output!")
		fmt.Println(out)
	}
	return out
}

func (c *CTemplateSet) compile_switch(varholder string, holdreflect reflect.Value, template_name string, node interface{}) (out string) {
	switch node := node.(type) {
		case *parse.ActionNode:
			if debug {
				fmt.Println("Action Node")
			}
			
			if node.Pipe == nil {
				break
			}
			for _, cmd := range node.Pipe.Cmds {
				out += c.compile_subswitch(varholder, holdreflect, template_name, cmd)
			}
			return out
		case *parse.IfNode:
			if debug {
				fmt.Println("If Node: ")
				fmt.Println(node.Pipe)
			}
			
			var expr string
			for _, cmd := range node.Pipe.Cmds {
				if debug {
					fmt.Println("If Node Bit: ")
					fmt.Println(cmd)
					fmt.Println(reflect.ValueOf(cmd).Type().Name())
				}
				expr += c.compile_varswitch(varholder, holdreflect, template_name, cmd)
			}
			
			if node.ElseList == nil {
				if debug {
					fmt.Println("Branch 1")
				}
				return "if " + expr + " {\n" + c.compile_switch(varholder, holdreflect, template_name, node.List) + "}\n"
			} else {
				if debug {
					fmt.Println("Branch 2")
				}
				return "if " + expr + " {\n" + c.compile_switch(varholder, holdreflect, template_name, node.List) + "} else {\n" + c.compile_switch(varholder, holdreflect, template_name, node.ElseList) + "}\n"
			}
		case *parse.ListNode:
			if debug {
				fmt.Println("List Node")
			}
			for _, subnode := range node.Nodes {
				out += c.compile_switch(varholder, holdreflect, template_name, subnode)
			}
			return out
		case *parse.RangeNode:
			if debug {
				fmt.Println("Range Node!")
				fmt.Println(node.Pipe)
			}
			
			var outVal reflect.Value
			for _, cmd := range node.Pipe.Cmds {
				if debug {
					fmt.Println("Range Bit: ")
					fmt.Println(cmd)
				}
				out, outVal = c.compile_reflectswitch(varholder, holdreflect, template_name, cmd)
			}
			
			if debug {
				fmt.Println("Returned: ")
				fmt.Println(out)
				fmt.Println("Range Kind Switch!")
			}
			
			switch outVal.Kind() {
				case reflect.Map:
					var item reflect.Value
					for _, key := range outVal.MapKeys() {
						item = outVal.MapIndex(key)
					}
					
					out = "if len(" + out + ") != 0 {\nfor _, item := range " + out + " {\n" + c.compile_switch("item", item, template_name, node.List) + "}\n}"
				case reflect.Invalid:
					return ""
			}
			
			if node.ElseList != nil {
				out += " else {\n" + c.compile_switch(varholder, holdreflect, template_name, node.ElseList) + "}\n"
			} else {
				out += "\n"
			}
			return out
		case *parse.TemplateNode:
			if debug {
				fmt.Println("Template Node")
			}
			return c.compile_subtemplate(varholder, holdreflect, node)
		case *parse.TextNode:
			return "w.Write([]byte(`" + string(node.Text) + "`))\n"
		default:
			panic("Unknown Node in main switch")
	}
	return ""
}

func (c *CTemplateSet) compile_subswitch(varholder string, holdreflect reflect.Value, template_name string, node *parse.CommandNode) (out string) {
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
		case *parse.FieldNode:
			if debug {
				fmt.Println("Field Node: ")
				fmt.Println(n.Ident)
			}
			
			/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
			cur := holdreflect
			
			var varbit string
			if cur.Kind() == reflect.Interface {
				cur = cur.Elem()
				varbit += ".(" + cur.Type().Name() + ")"
			}
			
			for _, id := range n.Ident {
				if debug {
					fmt.Println("Data Kind: ")
					fmt.Println(cur.Kind().String())
					fmt.Println("Field Bit: ")
					fmt.Println(id)
				}
				
				cur = cur.FieldByName(id)
				if cur.Kind() == reflect.Interface {
					cur = cur.Elem()
					/*if cur.Kind() == reflect.String && cur.Type().Name() != "string" {
						varbit = "string(" + varbit + "." + id + ")"*/
					//if cur.Kind() == reflect.String && cur.Type().Name() != "string" {
					if cur.Type().PkgPath() != "main" {
						c.importMap["html/template"] = "html/template"
						varbit += "." + id + ".(" + strings.TrimPrefix(cur.Type().PkgPath(),"html/") + "." + cur.Type().Name() + ")"
					} else {
						varbit += "." + id + ".(" + cur.Type().Name() + ")"
					}
				} else {
					varbit += "." + id
				}
				
				if debug {
					fmt.Println("End Cycle")
				}
			}
			out = c.compile_varsub(varholder + varbit, cur)
			
			for _, varItem := range c.varList {
				if strings.HasPrefix(out, varItem.Destination) {
					out = strings.Replace(out, varItem.Destination, varItem.Name, 1)
				}
			}
			return out
		case *parse.DotNode:
			if debug {
				fmt.Println("Dot Node")
				fmt.Println(node.String())
			}
			return c.compile_varsub(varholder, holdreflect)
		case *parse.NilNode:
			panic("Nil is not a command x.x")
		case *parse.VariableNode:
			if debug {
				fmt.Println("Variable Node")
				fmt.Println(n.String())
				fmt.Println(n.Ident)
			}
			
			out, _ = c.compile_if_varsub(n.String(), varholder, template_name, holdreflect)
			return "w.Write([]byte(" + out + "))\n"
		case *parse.StringNode:
			return n.Quoted
		default:
			fmt.Println("Unknown Kind: ")
			fmt.Println(reflect.ValueOf(firstWord).Elem().Kind())
			fmt.Println("Unknown Type: ")
			fmt.Println(reflect.ValueOf(firstWord).Elem().Type().Name())
			panic("I don't know what node this is")
	}
	return ""
}

func (c *CTemplateSet) compile_varswitch(varholder string, holdreflect reflect.Value, template_name string, node *parse.CommandNode) (out string) {
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
		case *parse.FieldNode:
			if debug {
				fmt.Println("Field Node: ")
				fmt.Println(n.Ident)
				
				for _, id := range n.Ident {
					fmt.Println("Field Bit: ")
					fmt.Println(id)
				}
			}
			
			/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
			return c.compile_boolsub(n.String(), varholder, template_name, holdreflect)
		case *parse.ChainNode:
			if debug {
				fmt.Println("Chain Node: ")
				fmt.Println(n.Node)
				fmt.Println(node.Args)
			}
			break
		case *parse.IdentifierNode:
			if debug {
				fmt.Println("Identifier Node: ")
				fmt.Println(node)
				fmt.Println(node.Args)
			}
			return c.compile_identswitch(varholder, holdreflect, template_name, node)
		case *parse.DotNode:
			return varholder
		case *parse.VariableNode:
			if debug {
				fmt.Println("Variable Node")
				fmt.Println(n.String())
				fmt.Println(n.Ident)
			}
			
			out, _ = c.compile_if_varsub(n.String(), varholder, template_name, holdreflect)
			return out
		case *parse.NilNode:
			panic("Nil is not a command x.x")
		case *parse.PipeNode:
			if debug {
				fmt.Println("Pipe Node!")
				fmt.Println(n)
			}
			
			/*for _, cmd := range n.Cmds {
				if debug {
					fmt.Println("Pipe Bit: ")
					fmt.Println(cmd)
				}
				out += c.compile_if_varsub_n(n.String(), varholder, template_name, holdreflect)
			}*/
			
			if debug {
				fmt.Println("Args: ")
				fmt.Println(node.Args)
			}
			
			/*argcopy := node.Args[1:]
			for _, arg := range argcopy {
				if debug {
					fmt.Println("Pipe Arg: ")
					fmt.Println(arg)
					fmt.Println(reflect.ValueOf(arg).Elem().Type().Name())
					fmt.Println(reflect.ValueOf(arg).Kind())
				}
				
				switch arg.(type) {
					case *parse.IdentifierNode:
						out += c.compile_identswitch(varholder, holdreflect, template_name, node)
						break
					case *parse.PipeNode:
						break
						//out += c.compile_if_varsub_n(a.String(), varholder, template_name, holdreflect)
					default:
						panic("Unknown Pipe Arg type! Did Mario get stuck in the pipes again?")
				}
				//out += c.compile_varswitch(arg.String(), holdreflect, template_name, arg)
			}*/
			out += c.compile_identswitch(varholder, holdreflect, template_name, node)
			
			if debug {
				fmt.Println("Out: ")
				fmt.Println(out)
			}
			return out
		default:
			fmt.Println("Unknown Kind: ")
			fmt.Println(reflect.ValueOf(firstWord).Elem().Kind())
			fmt.Println("Unknown Type: ")
			fmt.Println(reflect.ValueOf(firstWord).Elem().Type().Name())
			panic("I don't know what node this is! Grr...")
	}
	return ""
}

func (c *CTemplateSet) compile_identswitch(varholder string, holdreflect reflect.Value, template_name string, node *parse.CommandNode) (out string) {
	ArgLoop:
	for pos, id := range node.Args {
		if debug {
			fmt.Println(id)
		}
		
		switch id.String() {
			case "not":
				out += "!"
			case "or":
				out += " || "
			case "and":
				out += " && "
			case "le":
				out += c.compile_if_varsub_n(node.Args[pos + 1].String(), varholder, template_name, holdreflect) + " <= " + c.compile_if_varsub_n(node.Args[pos + 2].String(), varholder, template_name, holdreflect)
				break ArgLoop
			default:
				if debug {
					fmt.Println("Variable!")
				}
				out += c.compile_if_varsub_n(id.String(), varholder, template_name, holdreflect)
		}
	}
	return out
}

func (c *CTemplateSet) compile_reflectswitch(varholder string, holdreflect reflect.Value, template_name string, node *parse.CommandNode) (out string, outVal reflect.Value) {
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
		case *parse.FieldNode:
			if debug {
				fmt.Println("Field Node: ")
				fmt.Println(n.Ident)
				
				for _, id := range n.Ident {
					fmt.Println("Field Bit: ")
					fmt.Println(id)
				}
			}
			/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
			return c.compile_if_varsub(n.String(), varholder, template_name, holdreflect)
		case *parse.ChainNode:
			if debug {
				fmt.Println("Chain Node: ")
				fmt.Println(n.Node)
				fmt.Println(node.Args)
			}
			return "", outVal
		case *parse.DotNode:
			return varholder, holdreflect
		case *parse.NilNode:
			panic("Nil is not a command x.x")
		default:
			//panic("I don't know what node this is")
	}
	return "", outVal
}

func (c *CTemplateSet) compile_if_varsub_n(varname string, varholder string, template_name string, cur reflect.Value) (out string) {
	out, _ = c.compile_if_varsub(varname, varholder, template_name, cur)
	return out
}

func (c *CTemplateSet) compile_if_varsub(varname string, varholder string, template_name string, cur reflect.Value) (out string, val reflect.Value) {
	if varname[0] != '.' && varname[0] != '$' {
		return varname, cur
	}
	
	bits := strings.Split(varname,".")
	if varname[0] == '$' {
		var res VarItemReflect
		if varname[1] == '.' {
			res = c.localVars[template_name]["."]
		} else {
			res = c.localVars[template_name][strings.TrimPrefix(bits[0],"$")]
		}
		out += res.Destination
		cur = res.Value
		
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
		}
	} else {
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
			out += varholder + ".(" + cur.Type().Name() + ")"
		} else {
			out += varholder
		}
	}
	bits[0] = strings.TrimPrefix(bits[0],"$")
	
	if debug {
		fmt.Println("Cur Kind: ")
		fmt.Println(cur.Kind())
		fmt.Println("Cur Type: ")
		fmt.Println(cur.Type().Name())
	}
	
	for _, bit := range bits {
		if debug {
			fmt.Println("Variable Field!")
			fmt.Println(bit)
		}
		
		if bit == "" {
			continue
		}
		
		cur = cur.FieldByName(bit)
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
			out += "." + bit + ".(" + cur.Type().Name() + ")"
		} else {
			out += "." + bit
		}
		
		if debug {
			fmt.Println("Data Kind: ")
			fmt.Println(cur.Kind())
			fmt.Println("Data Type: ")
			fmt.Println(cur.Type().Name())
		}
	}
	
	for _, varItem := range c.varList {
		if strings.HasPrefix(out, varItem.Destination) {
			out = strings.Replace(out, varItem.Destination, varItem.Name, 1)
		}
	}
	
	_, ok := c.stats[out]
	if ok {
		c.stats[out]++
	} else {
		c.stats[out] = 1
	}
	
	return out, cur
}

func (c *CTemplateSet) compile_boolsub(varname string, varholder string, template_name string, val reflect.Value) string {
	out, val := c.compile_if_varsub(varname, varholder, template_name, val)
	switch val.Kind() {
		case reflect.Int:
			out += " > 0"
		case reflect.Bool:
			// Do nothing
		case reflect.String:
			out += " != \"\""
		case reflect.Int64:
			out += " > 0"
		default:
			panic("I don't know what this variable's type is o.o\n")
	}
	return out
}

func (c *CTemplateSet) compile_varsub(varname string, val reflect.Value) string {
	for _, varItem := range c.varList {
		if strings.HasPrefix(varname, varItem.Destination) {
			varname = strings.Replace(varname, varItem.Destination, varItem.Name, 1)
		}
	}
	
	_, ok := c.stats[varname]
	if ok {
		c.stats[varname]++
	} else {
		c.stats[varname] = 1
	}
	
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	
	switch val.Kind() {
		case reflect.Int:
			return "w.Write([]byte(strconv.Itoa(" + varname + ")))\n"
		case reflect.Bool:
			return "if " + varname + " {\nw.Write([]byte(\"true\"))} else {\nw.Write([]byte(\"false\"))\n}\n"
		case reflect.String:
			if val.Type().Name() != "string" && !strings.HasPrefix(varname,"string(") {
				return "w.Write([]byte(string(" + varname + ")))\n"
			} else {
				return "w.Write([]byte(" + varname + "))\n"
			}
		case reflect.Int64:
			return "w.Write([]byte(strconv.FormatInt(" + varname + ", 10)))"
		default:
			fmt.Println("Unknown Kind: ")
			fmt.Println(val.Kind())
			fmt.Println("Unknown Type: ")
			fmt.Println(val.Type().Name())
			panic("// I don't know what this variable's type is o.o\n")
	}
}

func (c *CTemplateSet) compile_subtemplate(pvarholder string, pholdreflect reflect.Value, node *parse.TemplateNode) (out string) {
	if debug {
		fmt.Println("Template Node: " + node.Name)
	}
	
	fname := strings.TrimSuffix(node.Name, filepath.Ext(node.Name))
	varholder := "tmpl_" + fname + "_vars"
	var holdreflect reflect.Value
	if node.Pipe != nil {
		for _, cmd := range node.Pipe.Cmds {
			firstWord := cmd.Args[0]
			switch firstWord.(type) {
				case *parse.DotNode:
					varholder = pvarholder
					holdreflect = pholdreflect
					break
				case *parse.NilNode:
					panic("Nil is not a command x.x")
				default:
					out = "var " + varholder + " := false\n"
					out += c.compile_command(cmd)
			}
		}
	}
	
	res, err := ioutil.ReadFile(c.dir + node.Name)
	if err != nil {
		log.Fatal(err)
	}
	content := string(res)
	
	tree := parse.New(node.Name, c.funcMap)
	var treeSet map[string]*parse.Tree = make(map[string]*parse.Tree)
	tree, err = tree.Parse(content,"{{","}}", treeSet, c.funcMap)
	if err != nil {
		log.Fatal(err)
	}
	
	c.tlist[fname] = tree
	subtree := c.tlist[fname]
	if debug {
		fmt.Println(subtree.Root)
	}
	
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".",varholder,holdreflect}
	
	for _, node := range subtree.Root.Nodes {
		if debug {
			fmt.Println("Node: " + node.String())
		}
		out += c.compile_switch(varholder, holdreflect, fname, node)
	}
	return out	
}

func (c *CTemplateSet) compile_command(*parse.CommandNode) (out string) {
	return ""
}