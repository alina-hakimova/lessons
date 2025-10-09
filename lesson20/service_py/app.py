from fastapi import FastAPI
import os

app = FastAPI(
    title="Orders Service",
    description="Microservice for handling orders", 
    version="v1"  
)

@app.get("/version")
async def get_version():
    """Get service version"""
    return {"version": os.getenv('VERSION', 'dev')}

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy", 
        "service": "orders",
        "version": os.getenv('VERSION', 'dev')
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "Orders Service v1"}  

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app:app", 
        host="0.0.0.0", 
        port=5000,
        reload=True if os.getenv('ENV') == 'development' else False
    )
