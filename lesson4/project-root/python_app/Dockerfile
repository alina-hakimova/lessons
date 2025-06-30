FROM python:3.11-slim
WORKDIR /app
COPY app.py .
RUN pip install flask
ARG UID
ARG GID
RUN addgroup --gid $GID appgroup && adduser --uid $UID --ingroup appgroup --disabled-password appuser
USER appuser
ENV PYTHONUNBUFFERED=1
EXPOSE 8001
HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -f http://localhost:8001/ || exit 1
ENTRYPOINT ["python", "app.py"]
