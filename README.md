# flexkids-photo-downloader
A command line tool to download photos from flexkids website.

Recently Kindergarden (https://www.kindergarden.nl/) informed us that they are going to remove
old photos of the kids from their parents portal. (https://kindergarden.flexkids.nl/)

Since my daughter has been going to Kindergarden for more than 2 years there are hundreds of photos of her.

As a lazy person instead of downloading these photos using their website I decided to create a simple tool.

## Usage

### Linux & Mac

 ./flexkids-photo-downloader -username username -password password -url url-of-the-flexkids-website -o output-directory

### Windows

 flexkids-photo-downloader-amd64.exe -username username -password password -url url-of-the-flexkids-website -o output-directory


  -o string
        output directory (default "output")

  -password string
        password

  -url string
        url of the flexkids web site (default "https://kindergarden.flexkids.nl")

  -username string
        username

You can download the windows executables from the bin directory. (I did not test if it works on windows)

Send me an email if you need some help. barane@gmail.com
