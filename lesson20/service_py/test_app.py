import pytest
from fastapi.testclient import TestClient
from app import app
import os

client = TestClient(app)

class TestOrdersService:
    def test_root_endpoint(self):
        """Test root endpoint returns correct message"""
        response = client.get("/")
        assert response.status_code == 200
        assert response.json() == {"message": "Orders Service v1"}
    
    def test_version_endpoint_default(self):
        """Test version endpoint without VERSION env variable"""
        response = client.get("/version")
        assert response.status_code == 200
        assert response.json() == {"version": "dev"}
    
    def test_version_endpoint_with_env(self, monkeypatch):
        """Test version endpoint with VERSION env variable"""
        monkeypatch.setenv('VERSION', 'git-abc123')
        response = client.get("/version")
        assert response.status_code == 200
        assert response.json() == {"version": "git-abc123"}
    
    def test_health_endpoint(self):
        """Test health endpoint returns healthy status"""
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "orders"
        assert "version" in data

    def test_nonexistent_endpoint(self):
        """Test 404 for nonexistent endpoint"""
        response = client.get("/nonexistent")
        assert response.status_code == 404
