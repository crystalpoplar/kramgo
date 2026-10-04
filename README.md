# kramgo

To execute build-deploy, stop the services on ubuntu, run this from repo root in PowerShell as admin:
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-upload-kramgo.ps1

To execute install in ubuntu, run:
sudo sed -i 's/\r$//' /home/dtk1376/install-kramgo-ollama.sh
sudo chmod 755 /home/dtk1376/install-kramgo-ollama.sh
sudo /home/dtk1376/install-kramgo-ollama.sh