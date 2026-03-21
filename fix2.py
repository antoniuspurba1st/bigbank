import pathlib
f = pathlib.Path(r"C:\Users\ronal\Desktop\com.bigbank\transaction-service\internal\handler\auth.go")
c = f.read_text(encoding="utf-8")
c = c.replace(""account_number": accountNumber,\n\t\t"balance":        balance,\n\t\t"transactions":   0,", ""id":             user.ID,\n\t\t"email":          user.Email,\n\t\t"phone":          user.Phone,\n\t\t"account_number": accountNumber,\n\t\t"balance":        balance,")
c = c.replace("writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": "Account not found"})", "writeError(w, correlationIDFromRequest(r), &model.AppError{\n\t\t\t\tStatusCode: http.StatusNotFound,\n\t\t\t\tCode:       "ACCOUNT_NOT_FOUND",\n\t\t\t\tMessage:    "Account not found",\n\t\t\t})")
c = c.replace(""Failed to create user: " + err.Error()", ""Failed to create user"")
f.write_text(c, encoding="utf-8")
print("done")
