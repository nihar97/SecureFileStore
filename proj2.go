package proj2


// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

//TODO: See if we need to marshall the data.

import (

	// You neet to add with
	// go get github.com/nweaver/cs161-p2/userlib
	"github.com/nweaver/cs161-p2/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging etc
	"encoding/hex"
	
	// UUIDs are generated right based on the crypto RNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// For the useful little debug printing function
	"fmt"
	"time"
	"os"
	"strings"

	// I/O
	"io"
	
	// Want to import errors
	"errors"
	
	// These are imported for the structure definitions.  You MUST
	// not actually call the functions however!!!
	// You should ONLY call the cryptographic functions in the
	// userlib, as for testing we may add monitoring functions.
	// IF you call functions in here directly, YOU WILL LOSE POINTS
	// EVEN IF YOUR CODE IS CORRECT!!!!!
	"crypto/rsa"
)


// This serves two purposes: It shows you some useful primitives and
// it suppresses warnings for items not being imported
func someUsefulThings(){
	// Creates a random UUID
	f := uuid.New()
	debugMsg("UUID as string:%v", f.String())
	
	// Example of writing over a byte of f
	f[0] = 10
	debugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	debugMsg("The hex: %v", h)
	
	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d,_ := json.Marshal(f)
	debugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	debugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	debugMsg("Creation of error %v", errors.New("This is an error"))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var key *rsa.PrivateKey
	key,_ = userlib.GenerateRSAKey()
	debugMsg("Key is %v", key)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range(ret){
		ret[x] = data[x]
	}
	return
}

// Helper function: Returns a byte slice of the specificed
// size filled with random data
func randomBytes(bytes int) (data []byte){
	data = make([]byte, bytes)
	if _, err := io.ReadFull(userlib.Reader, data); err != nil {
		panic(err)
	}
	return
}

var DebugPrint = false

// Helper function: Does formatted printing to stderr if
// the DebugPrint global is set.  All our testing ignores stderr,
// so feel free to use this for any sort of testing you want
func debugMsg(format string, args ...interface{}) {
	if DebugPrint{
		msg := fmt.Sprintf("%v ", time.Now().Format("15:04:05.00000"))
		fmt.Fprintf(os.Stderr,
			msg + strings.Trim(format, "\r\n ") + "\n", args...)
	}
}


// The structure definition for a user record
type User struct {
	Username string
	Password string
	RSAPrivKey *rsa.PrivateKey
	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}



// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored
// User data: the name used in the datastore should not be guessable
// without also knowing the password and username.

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the user has a STRONG password

//TODO: Double check filenaming scheme
func InitUser(username string, password string) (userdataptr *User, err error){
	var userdata User
	privkey, err := userlib.GenerateRSAKey()
	if err != nil {
		panic(err)
	}
	keys := userlib.PBKDF2Key(password, username, userlib.HashSize + userlib.AESKeySize)
	hmac_key := keys[0:userlib.HashSize]
	encrypt_key := keys[userlib.HashSize:]
	userdata = User{username, password, privkey}
	encrypted_data := EncryptData(encrypt_key, json.Marshal(userdata))
	hmac := GenerateHMAC(hmac_key, encrypted_data)
	packed_data := encrypted_data || hmac
	
	secure_filename := GenerateHMAC(hmac_key, username || password)
	userlib.DatastoreSet(secure_filename, packed_data)
	userlib.KeystoreSet(username, privkey.PublicKey)
	return &userdata, err
}



// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.

//TODO: Return an appropriate error message. Implement specificity of checks
func GetUser(username string, password string) (userdataptr *User, err error){
	var userdata User
	keys := userlib.PBKDF2Key(password, username, userlib.HashSize + userlib.AESKeySize)
	hmac_key := keys[0:userlib.HashSize]
	encrypt_key := keys[userlib.HashSize:]
	secure_filename := HMAC(hmac_key || username || password)
	packed_data, err := userlib.DatastoreGet(secure_filename)
	if (err) {
		panic("User data not found")
	}
	if len(packed_data < userlib.BlockSize) {
		panic("HMAC is invalid")
	}
	hmac := packed_data[len(packed_data) - userlib.BlockSize:]
	ciphertext := packed_data[0:len(packed_data) - userlib.BlockSize]
	if !VerifyHMAC(hmac_key, ciphertext, hmac) {
		panic("HMAC does not correspond to encrypted data")
	}
	plaintext = DecryptData(encrypt_key, ciphertext)
	userdata = json.Unmarshal(plaintext)
	return &userdata, err
}



// This stores a file in the datastore.
//
// The name of the file should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
}


// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.

func (userdata *User) AppendFile(filename string, data []byte) (err error){
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string)(data []byte, err error) {
	return
}

// You may want to define what you actually want to pass as a
// sharingRecord to serialized/deserialize in the data store.
type sharingRecord struct {
}


// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.

func (userdata *User) ShareFile(filename string, recipient string)(
	msgid string, err error){
	return 
}


// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	msgid string) error {
	return nil
}

// Removes access for all others.  
func (userdata *User) RevokeFile(filename string) (err error){
	return 
}

// Helper function encrypts data and returns ciphertext
func EncryptData(key byte[], plaintext []byte) (byte[]) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, userlib.BlockSize+len(plaintext))
	iv := ciphertext[userlib.BlockSize]
	if _, err := io.ReadFull(randomBytes(userlib.BlockSize), iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[userlib.BlockSize:], plaintext)
	return ciphertext
}

// Helper function verifies HMAC on ciphertext
func VerifyHMAC(key byte[], data byte[], old_mac byte[]) (bool) {
	new_mac := GenerateHMAC(key, data)
	return Equal(new_mac, old_mac)
}

// Helper function decrypts data and returns plaintext
func DecryptData(key byte[], ciphertext byte[]) (byte[]) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < userlib.BlockSize {
		//TODO: Handle this error appropriately
		panic("ciphertext too short")
	}
	iv := ciphertext[:userlib.BlockSize]
	ciphertext = ciphertext[userlib.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	return stream.XORKeyStream(ciphertext, ciphertext)
}

// Helper function takes data and returns MAC. Data must be encrypted.
func GenerateHMAC(key byte[], data byte[]) (byte[]) {
	mac := userlib.NewHMAC(key)
	mac.Write(data)
	mac_data := mac.Sum(nil)
	return mac_data
}

func ComputeMerkleRoot(byte[][] leaves) (byte[]) {
	for len(leaves) > 1 {
		hashes := make(byte[][])
		iter_count := len(leaves) / 2 + len(leaves) % 2
		for (iter_count > 0) {
			if len(leaves) == 1 {
				first, leaves := leaves[0], leaves[1:]
				hash := userlib.NewSHA256()
				hash.write(first)
				hash.Sum(nil)
				append(hashes, hash)
			} else {
				first, second, leaves := leaves[0], leaves[1], leaves[2:]
				hash := userlib.NewSHA256()
				hash.write(first || second)
				hash.Sum(nil)
				append(hashes, hash)
			}
		}
		leaves = hashes
	}
	return leaves[0]
}
