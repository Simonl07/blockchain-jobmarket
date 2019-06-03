package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	encoded_prefix []uint8
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	Db      map[string]Node
	Root    string
	Mapping map[string]string
}

// func (mpt *MerklePatriciaTrie) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(&struct {
// 		Db      map[string]Node   `json:"db"`
// 		Root    string            `json:"name"`
// 		Mapping map[string]string `mapping:"name"`
// 	}{
// 		Db:   mpt.Db,
// 		Root: mpt.Root,
// 	})
// }

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	path := toPath(key)
	result, error := getHelper(mpt, mpt.Root, path)
	return result, error
}

func getHelper(mpt *MerklePatriciaTrie, nodeHash string, path []uint8) (string, error) {
	node := mpt.Db[nodeHash]
	if isLeaf(node) {
		nibbles := compact_decode(node.flag_value.encoded_prefix)
		if eq(nibbles, path) {
			return node.flag_value.value, nil
		}
		return "", errors.New("path_not_found")
	} else if isExt(node) {
		nibbles := compact_decode(node.flag_value.encoded_prefix)
		restPath := path[len(common(nibbles, path)):]
		return getHelper(mpt, node.flag_value.value, restPath)
	} else {
		if len(path) == 0 {
			if node.branch_value[16] != "" {
				return node.branch_value[16], nil
			}
			return "", errors.New("path_not_found")
		}
		index := path[0]
		if node.branch_value[index] == "" {
			return "", errors.New("path_not_found")
		}
		return getHelper(mpt, node.branch_value[index], path[1:])
	}
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	mpt.Mapping[key] = new_value
	path := toPath(key)
	if len(mpt.Db) == 0 || mpt.Root == "" {
		//Adding first Root node
		_, hash := mpt.newNode(2, path, true, new_value)
		mpt.Root = hash
	} else {
		mpt.Root = insertHelper(mpt, mpt.Root, path, new_value)
	}
	return
}

func insertHelper(mpt *MerklePatriciaTrie, nodeHash string, path []uint8, newValue string) (output string) {
	node := mpt.Db[nodeHash]
	if isLeaf(node) {
		delete(mpt.Db, nodeHash)
		oldValue := node.flag_value.value

		nibbles := compact_decode(node.flag_value.encoded_prefix)
		commonPath := common(nibbles, path)
		restPath := path[len(commonPath):]
		restNibbles := nibbles[len(commonPath):]
		if len(restPath) == 0 && eq(commonPath, nibbles) {
			node.flag_value.value = newValue
			output = updateNode(mpt, node, nodeHash)
		} else {
			_, branchHash := mpt.newNode(1, nil, false, "")
			branchHash = insertHelper(mpt, branchHash, restPath, newValue)
			branchHash = insertHelper(mpt, branchHash, restNibbles, oldValue)
			if len(commonPath) == 0 {
				//L ->>> B
				output = branchHash
			} else {
				//L ->>> E -> B -> L
				_, extHash := mpt.newNode(2, commonPath, false, branchHash)
				output = extHash
			}
		}
	} else if isBranch(node) {
		if len(path) == 0 {
			node.branch_value[16] = newValue
			output = updateNode(mpt, node, nodeHash)
		} else {
			index := path[0]
			if node.branch_value[index] == "" {
				//No branch value here, add a new leaf with restPath
				_, leafHash := mpt.newNode(2, path[1:], true, newValue)
				node.branch_value[index] = leafHash
			} else {
				// The index already existed, call insert on what ever is under this index(branch, ext, leaf)
				node.branch_value[index] = insertHelper(mpt, node.branch_value[index], path[1:], newValue)
			}
			output = updateNode(mpt, node, nodeHash)
		}
	} else if isExt(node) {
		nibbles := compact_decode(node.flag_value.encoded_prefix)
		commonPath := common(nibbles, path)
		restPath := path[len(commonPath):]
		restNibbles := nibbles[len(commonPath):]
		if len(commonPath) != 0 {
			if len(restNibbles) == 0 {
				// Correct location found, insert into the branch node(index 16) that correspond with this ext node
				node.flag_value.value = insertHelper(mpt, node.flag_value.value, restPath, newValue)
			} else {
				/*
					Need to perform more insertion for remaining restPath, restNibbles, or both
					First record original branch node associated with current extension node,
					Modify current extension node's prefix to commonPath
					Create a new branch node & associated with current extension node.
					Now it is up to the following two situations(if statements) to make modifications to the new branch node
				*/
				oldBranchNodeHash := node.flag_value.value
				branchNode, branchNodeHash := mpt.newNode(1, nil, false, "")
				node.flag_value.encoded_prefix = compact_encode(nibbles[:len(commonPath)])
				if len(restNibbles) == 1 {
					// If only one nibble left, use a branch node instead of a second extension node,
					// directly wire the old branch node to new branch node
					branchNode.branch_value[restNibbles[0]] = oldBranchNodeHash
				} else {
					// If more than one nibble, use a new extension node to connect to oldBranchNode
					_, secondExtNodeHash := mpt.newNode(2, restNibbles[1:], false, oldBranchNodeHash)
					branchNode.branch_value[restNibbles[0]] = secondExtNodeHash
				}
				if len(restPath) != 0 {
					// If there is still remaining path for the new insertion, insert remaining path as a new leaf to the new branch node
					_, leafHash := mpt.newNode(2, restPath[1:], true, newValue)
					branchNode.branch_value[restPath[0]] = leafHash
				} else {
					// Insert new value into branch
					branchNode.branch_value[16] = newValue
				}
				branchNodeHash = updateNode(mpt, branchNode, branchNodeHash)
				node.flag_value.value = branchNodeHash
			}
			output = updateNode(mpt, node, nodeHash)
		} else {
			//No commonPath, convert E to B, add potential E for rest nibbles
			//E ->>> B -> E? ->> Rest
			oldBranchHash := node.flag_value.value
			delete(mpt.Db, nodeHash)
			branch, branchHash := mpt.newNode(1, nil, false, "")
			if len(restNibbles) != 0 {
				if len(restNibbles) > 1 {
					_, secondExtHash := mpt.newNode(2, restNibbles[1:], false, oldBranchHash)
					branch.branch_value[restNibbles[0]] = secondExtHash
				} else {
					branch.branch_value[restNibbles[0]] = oldBranchHash
				}
			} else {
				//Do nothing
			}
			if len(restPath) != 0 {
				_, leafHash := mpt.newNode(2, restPath[1:], true, newValue)
				branch.branch_value[restPath[0]] = leafHash
			} else {
				// Insert new value into branch
				branch.branch_value[16] = newValue
			}
			output = updateNode(mpt, branch, branchHash)
		}
	}
	return output
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	delete(mpt.Mapping, key)
	path := toPath(key)
	hash, _, err := del(mpt, mpt.Root, path)
	mpt.Root = hash
	if err != nil {
		return "", nil
	}
	return "", errors.New("path_not_found")
}

func del(mpt *MerklePatriciaTrie, nodeHash string, path []uint8) (string, bool, error) {
	node := mpt.Db[nodeHash]
	if isLeaf(node) {
		nibbles := compact_decode(node.flag_value.encoded_prefix)
		if eq(nibbles, path) {
			//FOUND
			delete(mpt.Db, nodeHash)
			return "old" + nodeHash, true, nil
		}
		return "", false, errors.New("path_not_found")
	} else if isExt(node) {
		nibbles := compact_decode(node.flag_value.encoded_prefix)
		commonPath := common(nibbles, path)
		restPath := path[len(commonPath):]
		hash, change, err := del(mpt, node.flag_value.value, restPath)
		if err == nil {
			node.flag_value.value = hash
		}
		nodeHash = updateNode(mpt, node, nodeHash)
		if change {
			delete(mpt.Db, nodeHash)
			childNode := mpt.Db[hash]
			_, newHash := mpt.newNode(2, append(nibbles, compact_decode(childNode.flag_value.encoded_prefix)...), isLeaf(childNode), childNode.flag_value.value)
			return newHash, false, nil
		}
		return nodeHash, false, nil
	} else {
		//Delete branch
		if len(path) == 0 {
			if node.branch_value[16] != "" {
				//Found the value
				node.branch_value[16] = ""
			} else {
				return "", false, errors.New("path_not_found")
			}
		} else {
			index := path[0]
			if node.branch_value[index] == "" {
				return "", false, errors.New("path_not_found")
			} else {
				hash, change, err := del(mpt, node.branch_value[index], path[1:])
				if err != nil {
					return updateNode(mpt, node, nodeHash), false, err
				}

				node.branch_value[index] = hash

				if change && strings.HasPrefix(hash, "old") {
					for i, v := range node.branch_value {
						if v == hash {
							node.branch_value[i] = ""
						}
					}
				}
			}
		}
		//re-adjust branch node
		brnches := branches(node)
		if len(brnches) == 1 {
			delete(mpt.Db, nodeHash)
			if node.branch_value[16] != "" {
				prefix := []uint8{}
				_, leafHash := mpt.newNode(2, prefix, true, node.branch_value[16])
				return leafHash, true, nil
			} else {
				index := brnches[0]
				childNodeHash := node.branch_value[index]
				childNode := mpt.Db[childNodeHash]
				if isBranch(childNode) {
					prefix := []uint8{uint8(index)}
					_, extHash := mpt.newNode(2, prefix, false, childNodeHash)
					return extHash, true, nil
				} else if isExt(childNode) {
					prefix := []uint8{uint8(index)}
					_, extHash := mpt.newNode(2, append(prefix, compact_decode(childNode.flag_value.encoded_prefix)...), false, childNode.flag_value.value)
					return extHash, true, nil
				} else {
					prefix := []uint8{uint8(index)}
					_, leafHash := mpt.newNode(2, append(prefix, compact_decode(childNode.flag_value.encoded_prefix)...), true, childNode.flag_value.value)
					return leafHash, true, nil
				}
			}
		} else {
			return updateNode(mpt, node, nodeHash), false, nil
		}
	}
}

func toPath(key string) []uint8 {
	bytes := []byte(key)
	path := []uint8{}
	for _, byte := range bytes {
		path = append(path, byte/16)
		path = append(path, byte%16)
	}
	return path
}

func eq(a, b []uint8) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (mpt *MerklePatriciaTrie) newNode(nodeType int, prefix []uint8, terminating bool, value string) (node Node, hash string) {
	var branches [17]string
	if nodeType == 1 {
		branches[16] = value
		node = Node{nodeType, branches, Flag_value{}}
	} else {
		if terminating {
			prefix = append(prefix, 16)
		}
		flags := Flag_value{compact_encode(prefix), value}
		node = Node{nodeType, branches, flags}
	}
	hash = node.hash_node()
	mpt.Db[hash] = node
	return node, hash
}

func updateNode(mpt *MerklePatriciaTrie, node Node, oldHash string) (newHash string) {
	delete(mpt.Db, oldHash)
	newHash = node.hash_node()
	mpt.Db[newHash] = node
	return
}

func common(a []uint8, b []uint8) []uint8 {
	diff := []uint8{}
	for i, v := range a {
		if i < len(b) && v == b[i] {
			diff = append(diff, v)
		} else {
			break
		}
	}
	return diff
}

func simpleHash(input string) string {
	if len(input) == 0 {
		return input
	}
	return input[len(input)-12 : len(input)-8]
}

func isLeaf(node Node) bool {
	if node.node_type == 2 && node.flag_value.encoded_prefix[0]/16 > 1 {
		return true
	}
	return false
}

func isBranch(node Node) bool {
	return node.node_type == 1
}

func isExt(node Node) bool {
	if node.node_type == 2 && node.flag_value.encoded_prefix[0]/16 < 2 {
		return true
	}
	return false
}

func TestCompact() {
	test_compact_encode()
}

func branches(node Node) []int {
	b := []int{}
	for i, v := range node.branch_value {
		if v != "" {
			b = append(b, i)
		}
	}
	return b
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

func compact_encode(hex_array []uint8) []uint8 {
	term := 0
	if hex_array[len(hex_array)-1] == 16 {
		term = 1
		hex_array = hex_array[:len(hex_array)-1]
	}
	oddlen := len(hex_array) % 2
	flags := uint8(2*term + oddlen)
	if oddlen == 1 {
		hex_array = append([]uint8{flags}, hex_array...)
	} else {
		hex_array = append([]uint8{flags, 0}, hex_array...)
	}
	output := []uint8{}
	for i := 0; i < len(hex_array); i += 2 {
		output = append(output, hex_array[i]*16+hex_array[i+1])
	}

	return output
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {

	temp := []uint8{}

	for _, n := range encoded_arr {
		temp = append(temp, n/16)
		temp = append(temp, n%16)
	}
	flag := temp[0]
	if flag%2 == 0 {
		temp = temp[2:]
	} else {
		temp = temp[1:]
	}
	return temp
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.Db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.Db[hash]))
	}
	return content
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value + string(compact_decode(node.flag_value.encoded_prefix))
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.Db = make(map[string]Node)
	mpt.Root = ""
	mpt.Mapping = make(map[string]string)
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func main() {

}
