# PulseCheck

Monitor HTTP simple escrito en Go. Lee una lista de URLs, ejecuta comprobaciones concurrentes y muestra estado, código HTTP y latencia.

## Uso

```bash
go run . urls.txt
go run . urls.txt --workers 10 --timeout 5s
go run . urls.txt --json
```

Cada línea de `urls.txt` debe contener una URL completa.
