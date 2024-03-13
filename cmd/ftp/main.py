import ftplib, sys, pathlib,argparse

FTP_HOST = 'ftp.channeladvisor.com'
FTP_USER = 'TESTPN:ftpuser@develin.com.au'
FTP_PASSWORD = 'Advisor2023!'

def list_ftp_directory(ftp_host, ftp_user, ftp_password, directory):
    # Connect to the FTP server
    ftp = ftplib.FTP(ftp_host)
    
    # Log in to the FTP server
    ftp.login(ftp_user, ftp_password)
    
    # Change to the desired directory
    ftp.cwd(directory)
    
    # List the files and directories in the current directory
    files = ftp.nlst()
    
    # Close the FTP connection
    ftp.quit()
    
    return files

def upload_file(ftp_host, ftp_user, ftp_password, file_path, target_path):
    # Connect to the FTP server
    ftp = ftplib.FTP(ftp_host)
    
    # Log in to the FTP server
    ftp.login(ftp_user, ftp_password)
    
    # Open the file to send
    with open(file_path, 'rb') as file:
        # Use FTP's STOR command to upload the file
        ftp.storbinary(f'STOR {target_path}', file)
    
    # Close the FTP connection
    ftp.quit()

def download_file(ftp_host, ftp_user, ftp_password, download_path, target_path):
    # Connect to the FTP server
    ftp = ftplib.FTP(ftp_host)
    
    # Log in to the FTP server
    ftp.login(ftp_user, ftp_password)
    
    # Open the file to send
    with open(download_path, 'wb') as file:
        # Use FTP's STOR command to upload the file
        ftp.retrbinary(f'RETR {target_path}', file.write)
    
    # Close the FTP connection
    ftp.quit()

def delete_file(ftp_host, ftp_user, ftp_password, target_path):
    # Connect to the FTP server
    ftp = ftplib.FTP(ftp_host)
    
    # Log in to the FTP server
    ftp.login(ftp_user, ftp_password)
    
    # Delte the file
    ftp.delete(target_path)
    
    # Close the FTP connection
    ftp.quit()

def main():
    parser = argparse.ArgumentParser(description='File operation script')
    parser.add_argument('cmd', choices=['upload', 'download', 'delete'], help='The command to execute')
    parser.add_argument('file_path', type=str, help='The file path to process')
    parser.add_argument('--upload_dir', type=str, default='InventoryUPLOAD/', help='The directory to upload to')
    parser.add_argument('--download_dir', type=str, default='./', help='The directory to download to')

    args = parser.parse_args()

    file_path = pathlib.Path(args.file_path)
    file_name = file_path.name
    file_path = str(file_path)
    cmd = args.cmd
    upload_dir = args.upload_dir
    download_dir = args.download_dir

    if cmd == 'upload':
        try:
            upload_file(FTP_HOST, FTP_USER, FTP_PASSWORD, file_path, upload_dir + file_name)
            print("ok")
        except Exception as e:
            sys.stderr.write("Error: " + str(e))
            sys.stderr.flush()
    elif cmd == 'download':
        try:
            download_file(FTP_HOST, FTP_USER, FTP_PASSWORD, file_path, download_dir + file_name)
            print("ok")
        except Exception as e:
            sys.stderr.write("Error: " + str(e))
            sys.stderr.flush()
    elif cmd == 'delete':
        try:
            delete_file(FTP_HOST, FTP_USER, FTP_PASSWORD, upload_dir+file_name)
            print("ok")
        except Exception as e:
            sys.stderr.write("Error: " + str(e))
            sys.stderr.flush()
    else:
        parser.print_usage()
        sys.exit(1)

if __name__ == "__main__":
    main()