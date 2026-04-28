# Administrative Panel UI form WasmForge

This is the UI for the administrative panel of WasmForge, 
an API Gateway platform for reverse-proxying and deploying
WASM plugins onto routes. This platform is designed to provide a 
user-friendly interface for managing and configuring the API 
Gateway, allowing users to easily deploy and manage their WASM plugins and access
rich visual statistical data.

## Workflow

It is designed to be used in conjunction with the WasmForge API Gateway,
as an embedded static web application. The UI is built using React and 
TypeScript, and is served as a static asset by the API Gateway. 

## Using as a standalone application

If you want to use the UI as a standalone application, you can follow these steps:

1. Clone the repository and navigate to the `ui/adminv2` directory.
2. Install the dependencies using `npm install`.
3. Start the development server using `npm start dev`. This will run the application on `http://localhost:3000`.
4. You can then access the UI in your web browser at `http://localhost:3000`.

This application will work natively in this mode, still accessing backend API endpoints.

## Using as an embedded static web application

To use the UI as an embedded static web application, you should follow
steps from the main README file of the WasmForge repository, which includes building the UI form and 
serving it as a static asset from the API Gateway.