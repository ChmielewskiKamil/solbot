### [{{ .Severity }}] {{ .Title }}

**File(s)**: {{ range .Locations }}[{{ .Position.Filename }}](link) {{ end }} 

**Description**: {{ .Description }}

**Recommendation(s)**: {{ .Recommendation }}

**Status**: Unresolved

**Update from the client**: 

---
