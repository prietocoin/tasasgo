# Imagen de TasasGo. El binario se compila en una etapa aparte y se ejecuta en
# una imagen vacía: el proceso no arrastra toolchain, gestor de paquetes ni
# glibc. Todo lo que necesita en tiempo de ejecución —certificados, usuario y
# zonas horarias— se resuelve aquí, porque el sistema operativo no existe.
#
#   docker build -t tasasgo .

# ---------- Etapa 1: compilación ----------
FROM golang:1.27-alpine AS builder

WORKDIR /src

# Las dependencias ocupan su propia capa: mientras no cambie go.mod, el rebuild
# reutiliza la capa y no vuelve a descargar el árbol de módulos.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO desactivado: el binario queda estático y no depende de glibc. Los símbolos
# se descartan porque nadie depura desde dentro de un contenedor sin shell.
# GOOS y GOARCH salen de los argumentos que inyecta el constructor y caen a
# linux/amd64 si no vienen: así el mismo Dockerfile compila desde Windows,
# desde un Mac ARM o desde la VPS sin editar una línea.
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go build -trimpath -ldflags="-s -w" -o /out/tasasgo ./cmd/server

# La imagen final no tiene /etc/passwd, así que las credenciales del usuario sin
# privilegios se fabrican aquí y se copian. Se escriben a mano en vez de usar
# adduser para que el uid y el gid sean exactamente 10001 y coincidan con el
# USER de abajo.
RUN printf 'tasasgo:x:10001:10001::/:/sbin/nologin\n' > /out/passwd && \
    printf 'tasasgo:x:10001:\n' > /out/group

# ---------- Etapa 2: ejecución ----------
FROM scratch

# Los certificados raíz son imprescindibles: el servicio habla por HTTPS con la
# API de Evolution y con el origen HOO. Sin ellos, TLS falla en cada llamada.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Sin estos dos archivos, USER 10001:10001 no tendría nombre y algunos clientes
# registran el proceso como "uid=10001" en lugar del servicio.
COPY --from=builder /out/passwd /etc/passwd
COPY --from=builder /out/group /etc/group

COPY --from=builder /out/tasasgo /tasasgo

# El panel viaja en la imagen: son el HTML, el manifiesto y las banderas que el
# servicio sirve desde el sistema de archivos, y son la última pieza que
# mantenía vivo al proyecto Node. El proceso arranca en la raíz y busca ./public.
WORKDIR /
COPY public /public

# El proceso no es root: un fallo explotable en el servicio no entrega el
# contenedor. Al no haber shell, tampoco hay escalada local que intentar.
USER 10001:10001

EXPOSE 3003

ENTRYPOINT ["/tasasgo"]
