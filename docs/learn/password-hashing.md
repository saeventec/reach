# Password hashing (argon2id)

## What it is

We never store a password. We store a **hash** of it: the result of a one-way function. The same password always gives the same hash, but the hash can't be turned back into the password.

At login, we hash what the person typed again, **with the same salt and settings** (read from the stored hash), and compare. If the hashes match, the password was right. That is `verifyPassword` in [services/auth/password_hash.go](../../services/auth/password_hash.go).

The comparison uses `subtle.ConstantTimeCompare`, which always takes the same time. A normal `==` stops at the first different byte, so it is a little faster when the first bytes differ. An attacker who measures that time could learn how many bytes were right. This is called a **timing attack**.

An analogy: a hash is like a fingerprint. You can check whether a finger matches a fingerprint, but you can't rebuild the finger from the print.

## Why it matters

Databases leak: a stolen backup, a bug, a wrong permission. If they hold passwords, every account is lost, and also the accounts on other sites where people reused the same password. With good hashes, an attacker has to **guess** each password and hash every guess.

So a password hash must be **slow and memory-hungry on purpose**. A normal hash like SHA-256 is built to be fast: a graphics card can try billions of guesses per second. argon2id makes each guess need 19 MiB of memory and some time. That is nothing for one login (about 25 ms), but very expensive for billions of guesses.

## Salt

A **salt** is a random value, new for every password, mixed in before hashing. Without it, two users with the same password would have the same hash, and an attacker could pre-compute the hashes of common passwords once and look them all up. With a salt, every hash must be attacked on its own. The salt is not secret: it is stored next to the hash.

## Tiny example

From [services/auth/password_hash.go](../../services/auth/password_hash.go):

```go
salt := make([]byte, 16)
rand.Read(salt) // 16 random bytes

hash := argon2.IDKey([]byte(password), salt, 2, 19*1024, 1, 32)
//                    password          salt  passes memory threads length
```

The stored result, in the standard **PHC string format**:

```
$argon2id$v=19$m=19456,t=2,p=1$k3Q0hZ8w...$Xq1vR7mN...
 algorithm version settings     salt         hash
```

The settings are stored with each hash. So we can raise them later: old hashes still verify with their old settings.

## The settings

We use **OWASP's recommended minimum**: 19 MiB of memory (`m=19456` KiB), 2 passes (`t=2`), 1 thread (`p=1`). The salt is 16 bytes and the hash 32 bytes, as [RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html) (the argon2 standard) recommends.

The library is `golang.org/x/crypto/argon2`. The `golang.org/x/...` packages are made by the Go team, but live outside the standard library, and argon2 is not in the standard library.

## The password rules

From [services/auth/password.go](../../services/auth/password.go) ([ADR 0025](../decisions/0025-password-rules-and-argon2id.md)):

- 8 to 128 characters. Characters, not bytes: `é` is 2 bytes but 1 character.
- Not the same as the username, and not containing "reach".
- Every character is allowed, in any language, also spaces. No "must contain a digit and a symbol" rules: NIST (the US standards body) forbids them, because they make passwords harder to remember, not harder to guess.

NIST's current rules ([SP 800-63B](https://pages.nist.gov/800-63-4/sp800-63b.html)) want at least 15 characters when the password is the only factor, as in Reach, and a check against a list of common or leaked passwords. We chose fewer rules for now (see the ADR).

## Unicode normalization

Some letters can be written in two ways: `é` is one code (U+00E9), or `e` plus a combining accent (U+0065 U+0301). They look the same, but they are different bytes, so they give different hashes. Different keyboards and systems produce different forms.

**NFC** ("Normalization Form C") is a standard rule that always picks the one-code form. We apply it to the password before checking and hashing (`norm.NFC.String`, from `golang.org/x/text`), as NIST asks. So the same password works on every device.

## Alternatives

- **bcrypt:** older and widely used. Only uses the first 72 bytes of a password, and needs little memory, so graphics cards attack it more easily.
- **scrypt:** memory-hungry like argon2, older, with harder-to-pick settings.
- **PBKDF2:** fast with graphics cards unless you use a very high pass count. Only use it when a standard requires it.
- **A plain hash like SHA-256:** never for passwords: much too fast.

## Learn more

- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [NIST SP 800-63B: password rules](https://pages.nist.gov/800-63-4/sp800-63b.html)
- [The Go argon2 package](https://pkg.go.dev/golang.org/x/crypto/argon2)
