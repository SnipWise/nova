sur le même principe que ChunkXML, je voudrais une nouvelle fonction pour chunker du YAML.

comme pour ChunkXML(xml string, targetTag string) []string 
je voudrais une fonction ChunkYAML(yaml string, targetKey string) []string qui prend une chaîne YAML et une clé cible, et retourne une liste de chaînes YAML chunkées en fonction de la clé cible.

```yaml
snippets:
  snippet:
    name: "Chunk YAML"
    description: "Chunk YAML content into smaller pieces for easier processing."
    language: "python"
    code: |
        import yaml

        def chunk_yaml(yaml_content, chunk_size):
            """
            Chunk YAML content into smaller pieces.

            Parameters:
            yaml_content (str): The YAML content to be chunked.
            chunk_size (int): The maximum size of each chunk.

            Returns:
            list: A list of YAML chunks.
            """
            # Load the YAML content into a Python object
            data = yaml.safe_load(yaml_content)

            # Convert the Python object back to a YAML string
            yaml_string = yaml.dump(data)

            # Split the YAML string into chunks
            chunks = [yaml_string[i:i + chunk_size] for i in range(0, len(yaml_string), chunk_size)]

        return chunks  
  snippet:
    name: "Example Usage"
    description: "Example of how to use the chunk_yaml function."
    language: "python"
    code: |
        yaml_content = """
        name: John Doe
        age: 30
        address:
          street: 123 Main St
          city: Anytown
          state: CA
          zip: 12345
        """

        chunk_size = 50
        chunks = chunk_yaml(yaml_content, chunk_size)

        for i, chunk in enumerate(chunks):
            print(f"Chunk {i + 1}:\n{chunk}\n")
```

la clé cible paut aussi être de ce type: `- id`:

```yaml
snippets:
  - id: 1
    name: hello_world
    description: "Print Hello World to the console"
    language: swiftlang
    code: |
      // The simplest Swift program
      print("Hello, World!")

  - id: 2
    name: variables_and_constants
    description: "Declare variables and constants"
    language: swiftlang
    code: |
      // Variable (mutable)
      var name = "Alice"
      var age: Int = 30
      name = "Bob" // OK, variables can be reassigned

      // Constant (immutable)
      let pi = 3.14159
      let appName: String = "MyApp"
      // pi = 3.0 // Error! let is immutable

      // Multiple declarations
      var width = 100, height = 200

      // Optional (can be nil)
      var email: String? = nil
      email = "alice@example.com"

      print(name, age, pi, appName, width, height, email ?? "no email")
```

ensuite tu creeras un exemple dans /samples/200-rag-with-yaml en utilisant le fichier yaml present dans le dossier

puis tu mettras à jour la documentation pour inclure cette nouvelle fonction de chunking YAML, en expliquant comment elle fonctionne et en fournissant des exemples d'utilisation.
Dans /docs/rag-agent-guide-en.md et /docs/rag-agent-guide-fr.md, tu ajouteras une section dédiée à la fonction ChunkYAML, en détaillant son utilité, son fonctionnement et en fournissant des exemples concrets d'utilisation.


