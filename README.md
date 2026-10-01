# PulseCheck — Monitor HTTP concurrente

Comprueba múltiples URLs en paralelo, mide latencia y muestra rápidamente qué servicios responden y cuáles fallan.

## Uso

```bash
go run . urls.txt
go run . urls.txt --workers 10 --timeout 5s
go run . urls.txt --json
```

Cada línea de `urls.txt` debe contener una URL completa.
