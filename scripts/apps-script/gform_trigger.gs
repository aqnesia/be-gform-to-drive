/**
 * Google Apps Script - Trigger onFormSubmit
 * Pasang script ini pada container Google Form atau Google Sheet tanggapan.
 * Set trigger: onFormSubmit
 */

const BACKEND_URL = "https://your-backend-domain.com/api/webhook/gform";
const WEBHOOK_SECRET = "your-secret-token-here";

function onFormSubmit(e) {
  if (!e) {
    Logger.log("No event object found");
    return;
  }

  const itemResponses = e.response.getItemResponses();
  const formData = {};
  let cvFileBase64 = "";
  let cvFileName = "";

  for (let i = 0; i < itemResponses.length; i++) {
    const itemResponse = itemResponses[i];
    const item = itemResponse.getItem();
    const title = item.getTitle();
    const response = itemResponse.getResponse();

    // Periksa apakah item ini merupakan upload file CV
    if (item.getType() === FormApp.ItemType.FILE_UPLOAD) {
      try {
        const fileIds = response; // Array of Drive File IDs
        if (fileIds && fileIds.length > 0) {
          const file = DriveApp.getFileById(fileIds[0]);
          cvFileName = file.getName();
          cvFileBase64 = Utilities.base64Encode(file.getBlob().getBytes());
        }
      } catch (err) {
        Logger.log("Gagal membaca file upload: " + err.toString());
      }
    } else {
      formData[title] = response;
    }
  }

  // Waktu submit dari Google Form (digunakan sebagai 作成日 / Tanggal Pembuatan)
  const submitTime = e.response.getTimestamp();

  const payload = {
    response_id: e.response.getId(),
    submit_time: submitTime.toISOString(),
    cv_file_name: cvFileName,
    cv_file_base64: cvFileBase64,
    data: formData
  };

  const options = {
    method: "post",
    contentType: "application/json",
    headers: {
      "X-Webhook-Secret": WEBHOOK_SECRET
    },
    payload: JSON.stringify(payload),
    muteHttpExceptions: true
  };

  try {
    const res = UrlFetchApp.fetch(BACKEND_URL, options);
    Logger.log("Response Code: " + res.getResponseCode());
    Logger.log("Response Body: " + res.getContentText());
  } catch (err) {
    Logger.log("Error sending to backend: " + err.toString());
  }
}
