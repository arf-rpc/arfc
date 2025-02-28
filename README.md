# arfc

**arfc** is the _arf compiler_/transpiler. It is responsible for transforming 
IDL definitions into definitions usable by other languages.

For information on how to use IDL to define arf structures, services and 
clients, see the [idl](https://github.com/arf-rpc/idl) repository.

## Usage
```
arfc -l LANG -i INPUT -o OUTPUT

--input value, -i value   Input IDL file to be used to generate sources 
--lang value, -l value    The target language (go/golang/ruby)
--output value, -o value  Directory path to emit sources to
--ruby-module value       When lang is set to "ruby", overrides the module in 
                          which generated sources will be contained within. 
                          Must be in the format Some::Module::Path.

```

Options `--input`, `--lang`, and `--output` are required.

- `--input` (or `-i`) takes the path to an IDL file containing definitions to be generated to the target language.
- `--output` (or `-o`) takes the path to the destination directory where source files will be written.
- `--lang` (or `-l`) takes the target language in which the tool will generate sources.

When generating sources to `ruby`, the following options are available:

- `--ruby-module`: Overrides the module path in which classes will be generated. By default, the tool takes the `package` value of the input IDL and converts it into a module path. When defined, this overrides the detected value from the IDL file.
