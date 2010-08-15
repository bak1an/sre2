
package sre2

import (
  "fmt"
  "testing"
)

// Check the given state to be true.
func checkState(t *testing.T, state bool, err string) {
  if !state {
    t.Error(err)
  }
}

// Check the equality of two []int slices.
func checkIntSlice(t *testing.T, expected []int, result []int, err string) {
  match := true
  if len(expected) != len(result) {
    match = false
  } else {
    for i := 0; i < len(expected); i++ {
      if expected[i] != result[i] {
        match = false
      }
    }
  }
  checkState(t, match, fmt.Sprintf("%s: got %s, expected %s", err, result, expected))
}

// Run a selection of basic regular expressions against this package.
func TestSimpleRe(t *testing.T) {
  r := MustParse("")
  checkState(t, r.RunSimple(""), "everything should match")
  checkState(t, r.RunSimple("fadsnjkflsdafnas"), "everything should match")

  r = MustParse("^(a|b)+c*$")
  checkState(t, !r.RunSimple("abd"), "not a valid match")
  checkState(t, r.RunSimple("a"), "basic string should match")
  checkState(t, !r.RunSimple(""), "empty string should not match")
  checkState(t, r.RunSimple("abcccc"), "longer string should match")

  r = MustParse("(\\w*)\\s*(\\w*)")
  ok, res := r.RunSubMatch("zing hello there")
  checkState(t, ok, "should match generally")
  checkIntSlice(t, []int{0, 10, 0, 4, 5, 10}, res, "did not match first two words as expected")

  r = MustParse(".*?(\\w+)$")
  ok, res = r.RunSubMatch("zing hello there")
  checkState(t, ok, "should match generally")
  checkIntSlice(t, []int{0, 16, 11, 16}, res, "did not match last word as expected")
}

// Test parsing an invalid RE returns an error.
func TestInvalidRe(t *testing.T) {
  r, err := Parse("a**")
  checkState(t, err != nil, "must fail parsing")
  checkState(t, r == nil, "regexp must be nil")

  pass := false
  func() {
    defer func() {
      if r := recover(); r != nil {
        pass = true
      }
    }()
    MustParse("z(((a")
  }()
  checkState(t, pass, "should panic")
}

// Test behaviour related to character classes expressed within [...].
func TestCharClass(t *testing.T) {
  r := MustParse("^[\t[:word:]]+$") // Match tabs and word characters.
  checkState(t, r.RunSimple("c"), "non-space should match")
  checkState(t, !r.RunSimple("c t"), "space should not match")
  checkState(t, r.RunSimple("c\tt"), "tab should match")

  r = MustParse("^[:ascii:]*$")
  checkState(t, r.RunSimple(""), "nothing should match")
  checkState(t, r.RunSimple("c"), "ascii should match")
  checkState(t, !r.RunSimple("Π"), "unicode should not match")

  r = MustParse("^\\pN$")
  checkState(t, r.RunSimple("〩"), "character from Nl should match")
  checkState(t, r.RunSimple("¾"), "character from Nu should match")

  r = MustParse("^\\p{Nl}$")
  checkState(t, r.RunSimple("〩"), "character from Nl should match")
  checkState(t, !r.RunSimple("¾"), "character from Nu should not match")

  r = MustParse("^[^. ]$")
  checkState(t, r.RunSimple("\n"), "should match \\n")
  checkState(t, !r.RunSimple(" "), "should match only \\n")

  r = MustParse("^[.\n]$")
  checkState(t, r.RunSimple("\n"), "should match \\n")

  r = MustParse("^\\W$")
  checkState(t, !r.RunSimple("a"), "should not match word char")
  checkState(t, r.RunSimple("!"), "should match non-word")

  r = MustParse("^[abc\\W]$")
  checkState(t, r.RunSimple("a"), "should match 'a'")
  checkState(t, r.RunSimple("!"), "should match '!'")
  checkState(t, !r.RunSimple("d"), "should not match 'd'")

  r = MustParse("^[^abc\\W]$")
  checkState(t, !r.RunSimple("a"), "should not match 'a'")
  checkState(t, !r.RunSimple("%"), "should not match non-word char")
  checkState(t, r.RunSimple("d"), "should match 'd'")

  r = MustParse("^[\\w\\D]$")
  checkState(t, r.RunSimple("a"), "should match regular char 'a'")
  checkState(t, r.RunSimple("2"), "should still match number '2', caught by \\w")
}

// Test regexp generated by escape sequences (e.g. \n, \. etc).
func TestEscapeSequences(t *testing.T) {
  r := MustParse("^\\.\n\\044$") // Match '.\n$'
  checkState(t, r.RunSimple(".\n$"), "should match")
  checkState(t, !r.RunSimple(" \n$"), "space should not match")
  checkState(t, !r.RunSimple("\n\n$"), ". does not match \n by default")
  checkState(t, !r.RunSimple(".\n"), "# should not be treated as end char")

  r = MustParse("^\\x{03a0}\\x25$") // Match 'Π%'.
  checkState(t, r.RunSimple("Π%"), "should match pi+percent")

  r, err := Parse("^\\Π$")
  checkState(t, err != nil && r == nil,
      "should have failed on trying to escape Π, not punctuation")
}

// Tests string literals between \Q...\E.
func TestStringLiteral(t *testing.T) {
  r := MustParse("^\\Qhello\\E$")
  checkState(t, r.RunSimple("hello"), "should match hello")

  r = MustParse("^\\Q.$\\\\E$") // match ".$\\"
  checkState(t, r.RunSimple(".$\\"), "should match")
  checkState(t, !r.RunSimple(" $\\"), "should not match")

  r = MustParse("^a\\Q\\E*b$") // match absolutely nothing between 'ab'
  checkState(t, r.RunSimple("ab"), "should match")
  checkState(t, !r.RunSimple("acb"), "should not match")
}

// Test closure expansion types, such as {..}, ?, +, * etc.
func TestClosureExpansion(t *testing.T) {
  r := MustParse("^za?$")
  checkState(t, r.RunSimple("z"), "should match none")
  checkState(t, r.RunSimple("za"), "should match single")
  checkState(t, !r.RunSimple("zaa"), "should not match more")

  r = MustParse("^a{2,2}$")
  checkState(t, !r.RunSimple(""), "0 should fail")
  checkState(t, !r.RunSimple("a"), "1 should fail")
  checkState(t, r.RunSimple("aa"), "2 should succeed")
  checkState(t, r.RunSimple("aaa"), "3 should succeed")
  checkState(t, r.RunSimple("aaaa"), "4 should succeed")
  checkState(t, !r.RunSimple("aaaaa"), "5 should fail")

  r = MustParse("^a{2}$")
  checkState(t, !r.RunSimple(""), "0 should fail")
  checkState(t, !r.RunSimple("a"), "1 should fail")
  checkState(t, r.RunSimple("aa"), "2 should succeed")
  checkState(t, !r.RunSimple("aaa"), "3 should fail")

  r = MustParse("^a{3,}$")
  checkState(t, !r.RunSimple("aa"), "2 should fail")
  checkState(t, r.RunSimple("aaa"), "3 should succeed")
  checkState(t, r.RunSimple("aaaaaa"), "more should succeed")
}

// Test specific greedy/non-greedy closure types.
func TestClosureGreedy(t *testing.T) {
  r := MustParse("^(a{0,2}?)(a*)$")
  ok, res := r.RunSubMatch("aaa")
  checkState(t, ok, "should match")
  checkIntSlice(t, []int{0, 3, 0, 0, 0, 3}, res, "did not match expected")

  r = MustParse("^(a{0,2})?(a*)$")
  ok, res = r.RunSubMatch("aaa")
  checkState(t, ok, "should match")
  checkIntSlice(t, []int{0, 3, 0, 2, 2, 3}, res, "did not match expected")

  r = MustParse("^(a{2,}?)(a*)$")
  ok, res = r.RunSubMatch("aaa")
  checkState(t, ok, "should match")
  checkIntSlice(t, []int{0, 3, 0, 2, 2, 3}, res, "did not match expected")
}

// Test simple left/right matchers.
func TestLeftRight(t *testing.T) {
  r := MustParse("^.\\b.$")
  checkState(t, r.RunSimple("a "), "left char is word")
  checkState(t, r.RunSimple(" a"), "right char is word")
  checkState(t, !r.RunSimple("  "), "not a boundary")
  checkState(t, !r.RunSimple("aa"), "not a boundary")
}

// Test general flags in sre2.
func TestFlags(t *testing.T) {
  r := MustParse("^(?i:AbC)zz$")
  checkState(t, r.RunSimple("abczz"), "success")
  checkState(t, !r.RunSimple("abcZZ"), "fail, flag should not escape")
  ok, res := r.RunSubMatch("ABCzz")
  checkState(t, ok, "should pass")
  checkIntSlice(t, []int{0,5}, res, "should just have a single outer paren")

  r = MustParse("^(?U)(a+)(.+)$")
  ok, res = r.RunSubMatch("aaaabb")
  checkState(t, ok, "should pass")
  checkIntSlice(t, []int{0,6,0,1,1,6}, res, "should be ungreedy")

  r = MustParse("^(?i)a*(?-i)b*$")
  checkState(t, r.RunSimple("AAaaAAaabbbbb"), "success")
  checkState(t, !r.RunSimple("AAaaAAaaBBBa"), "should fail, flag should not escape")
}

// Test the behaviour of rune filters.
func TestRuneFilter(t *testing.T) {
  var filter RuneFilter

  filter = MatchRune('#')
  checkState(t, !filter('B'), "should not match random rune")
  checkState(t, filter('#'), "should match configured rune")

  filter = MatchRuneRange('A', 'Z')
  checkState(t, filter('A'), "should match rune 'A' in range")
  checkState(t, filter('B'), "should match rune 'B' in range")

  filter = MatchUnicodeClass("Greek")
  checkState(t, filter('Ω'), "should match omega")
  checkState(t, !filter('Z'), "should not match regular latin rune")

  filter = MatchUnicodeClass("Cyrillic").Not()
  checkState(t, filter('%'), "should match a random non-Cyrillic rune")
  checkState(t, !filter('Ӄ'), "should not match Cyrillic rune")
}

// Test the SafeParser used by much of the code.
func TestStringParser(t *testing.T) {
  src := NewSafeReader("a{bc}d")

  checkState(t, src.curr() == -1, "should not yet be parsing")
  checkState(t, src.nextCh() == 'a', "first char should be a")
  checkState(t, src.nextCh() == '{', "second char should be {")
  lit := src.literal("{", "}")
  checkState(t, lit == "bc", "should equal contained value, got: " + lit)
  checkState(t, src.curr() == 'd', "should now rest on d")
  checkState(t, src.nextCh() == -1, "should be done now")
}
