package helpers

// House struct holds scraped data
// Pointers denote that data is not from the JSON and can be empty
type House struct {
	Property struct {
		DatePostedString       string `json:"datePostedString,omitempty"`
		Description            string `json:"description,omitempty"`
		DesktopWebHdpImageLink string `json:"desktopWebHdpImageLink,omitempty"`
		HdpUrl                 string `json:"hdpUrl,omitempty"`
		HomeStatus             string `json:"homeStatus,omitempty"`
		HomeType               string `json:"homeType,omitempty"`

		// Non-JSON fields (manually manipulated)
		Address          string `json:"address,omitempty"`
		FullUrl          string `json:"fullUrl,omitempty"`
		MapsUrl          string `json:"mapsURL,omitempty"`
		PriceDiff        int    `json:"priceDiff",omitempty`
		PriceDiffPercent int    `json:"priceDiffPercent,omitempty"`
		ListPrice        int    `json:"listPrice,omitempty"`
		ListDate         string `json:"listDate,omitempty"`
		SoldPrice        int    `json:"soldPrice,omitempty"`
		SoldDate         string `json:"soldDate,omitempty"`

		ResoFacts struct {
			Bathrooms  int    `json:"bathrooms,omitempty"`
			Bedrooms   int    `json:"bedrooms,omitempty"`
			LivingArea string `json:"livingArea,omitempty"`
		} `json:"resoFacts,omitempty"`

		OpenHouseSchedule []struct {
			StartTime string `json:"startTime,omitempty"`
			EndTime   string `json:"endTime,omitempty"`
		} `json:"openHouseSchedule,omitempty"`

		PriceHistory []struct {
			Date  string `json:"date,omitempty"`
			Event string `json:"event,omitempty"`
			Price int    `json:"price,omitempty"`
		} `json:"priceHistory,omitempty"`
	} `json:"property,omitempty"`
}

// HouseSlice is a slice of House
type HouseSlice []House

// Details struct holds arguments for queries
type Details struct {
	Beds, Baths, Price     int
	Location, PropertyType string
	Sold                   bool
}

const TemplateHTML = `
<!DOCTYPE html>
<html>
<head>
<script src="https://cdnjs.cloudflare.com/ajax/libs/tablesort/5.1.0/tablesort.min.js"></script>
<title>Property Listings in "{{.Location}}" @ {{.At}}</title>
<style>
input#myInput { width: 220px; }
table#data-table {width: 100%;}
th, td {border: 1px solid #ddd; padding: 8px; text-align: left;}
th {background-color: #f1f1f1; cursor: pointer;}
tr:nth-child(even) {background-color: #f2f2f2}
</style>
</head>
<body>
{{- if .Sold}}
<h1>Sold Property Listings in "{{.Location}}" @ {{.At}}</h1>
{{- else}}
<h1>For Sale Property Listings in "{{.Location}}" @ {{.At}}</h1>
{{- end}}
<input
  type="text"
  id="myInput"
  onkeyup="myFunction()"
  placeholder="Type to filter results..."
  title="Type in anything to begin filtering...">
<table id="data-table">
  <thead>
    <tr data-sort-method="none">
      <th>Image</th>
      <th>Address</th>
      <th>Beds</th>
      <th>Baths</th>
      <th>Description</th>
      <th>Google Maps</th>
      <th>Showing?</th>
      <th>Size</th>
      <th>List Date</th>
      <th>List Price</th>
      <th>Sold Date</th>
      <th>Sold Price</th>
      <th>Price Diff</th>
      <th>Status</th>
      <th>Property Type</th>
    </tr>
  </thead>
  <tbody id="table-body">
    {{range .Houses}}
      <tr>
        <td><img src="{{.Img}}" alt="House Image" style="width: 100px; height: 100px;"></td>
        <td><a href={{.Link}}>{{.Address}}</a></td>
        <td>{{.Beds}}</td>
        <td>{{.Baths}}</td>
        <td>{{.Description}}</td>
        <td><a href={{.MapsURL}}>Link</a></td>
        <td>{{.Showing}}</td>
        <td>{{.Size}}</td>
        <td>{{.ListDate}}</td>
        <td>{{.ListPrice}}</td>
        <td>{{.SoldDate}}</td>
        <td>{{.SoldPrice}}</td>
        <td>{{.PriceDiff}}</td>
        <td>{{.Status}}</td>
        <td>{{.PropertyType}}</td>
      </tr>
    {{end}}
  </tbody>
</table>
<script>
function onPageReady() {
  // Documentation: http://tristen.ca/tablesort/demo/
  new Tablesort(document.getElementById('data-table'));
}

// Run the above function when the page is loaded & ready
document.addEventListener('DOMContentLoaded', onPageReady, false);

const myFunction = () => {
  const trs = document.querySelectorAll('#data-table tr'); // Include all rows
  const filter = document.querySelector('#myInput').value;
  const regex = new RegExp(filter, 'i');

  const isFoundInTds = (td) => regex.test(td.textContent); // Use textContent for header
  const isFound = (childrenArr) => childrenArr.some(isFoundInTds);

  const setTrStyleDisplay = ({ style, children }) => {
    if (children[0].tagName === 'TH') { // If header row
      style.display = ''; // Always show header
    } else {
      style.display = isFound([...children]) ? '' : 'none';
    }
  };

  trs.forEach(setTrStyleDisplay);
};

</script>
</body>
</html>
`
