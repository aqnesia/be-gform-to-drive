/**
 * Google Apps Script - Web App Upload Drive
 * Deploy as Web App:
 * - Execute as: Me (your Google account)
 * - Who has access: Anyone
 */

const DRIVE_SECRET = "your-drive-script-secret-here";

function doPost(e) {
  try {
    const authHeader = e.parameter.secret || (e.postData && JSON.parse(e.postData.contents).secret);
    if (authHeader !== DRIVE_SECRET) {
      return ContentService.createTextOutput(JSON.stringify({
        status: "error",
        message: "Unauthorized"
      })).setMimeType(ContentService.MimeType.JSON);
    }

    const body = JSON.parse(e.postData.contents);
    const fileName = body.file_name;
    const fileBytes = Utilities.base64Decode(body.file_base64);
    const blob = Utilities.newBlob(fileBytes, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileName);

    // Dapatkan atau buat folder: CV / YYYY-MM
    const date = new Date();
    const yearMonth = Utilities.formatDate(date, Session.getScriptTimeZone(), "yyyy-MM");
    
    let rootCVFolder;
    const rootFolders = DriveApp.getFoldersByName("CV");
    if (rootFolders.hasNext()) {
      rootCVFolder = rootFolders.next();
    } else {
      rootCVFolder = DriveApp.createFolder("CV");
    }

    let targetFolder;
    const subFolders = rootCVFolder.getFoldersByName(yearMonth);
    if (subFolders.hasNext()) {
      targetFolder = subFolders.next();
    } else {
      targetFolder = rootCVFolder.createFolder(yearMonth);
    }

    // Periksa nama kembar
    let finalFileName = fileName;
    let counter = 1;
    while (targetFolder.getFilesByName(finalFileName).hasNext()) {
      counter++;
      const nameWithoutExt = fileName.replace(/\.xlsx$/i, "");
      finalFileName = nameWithoutExt + "_" + counter + ".xlsx";
    }

    blob.setName(finalFileName);
    const createdFile = targetFolder.createFile(blob);

    return ContentService.createTextOutput(JSON.stringify({
      status: "success",
      file_id: createdFile.getId(),
      file_url: createdFile.getUrl(),
      file_name: finalFileName
    })).setMimeType(ContentService.MimeType.JSON);

  } catch (error) {
    return ContentService.createTextOutput(JSON.stringify({
      status: "error",
      message: error.toString()
    })).setMimeType(ContentService.MimeType.JSON);
  }
}

