load 'helpers/bats-support/load'
load 'helpers/bats-assert/load'

setup() {
  export TESTFILE="testfile_1.txt"
  # Create a temporary file with some content
  echo "This is a test file." > $TESTFILE
  #run gcloud storage cp $TESTFILE gs://$BUCKET/$TESTFILE
}

teardown() {
  # Remove the temporary file
  rm $TESTFILE
  #run gcloud storage rm gs://$BUCKET/$TESTFILE
}

@test "Test gcloud storage cp" {
  run gcloud storage cp $TESTFILE gs://$BUCKET/$TESTFILE
  assert_success
}

@test "Test gcloud storage cat - returns exact object" {
  local expected_output=$(cat $TESTFILE)
  run gcloud storage cat gs://$BUCKET/$TESTFILE
  assert_success
  assert_output "$expected_output"  
}

@test "Test gcloud storage cat - returns correct byte length" {
  # Get the expected byte length (replace with your actual logic)
  local expected_length=$(wc -c $TESTFILE)

  # Run the gcloud command and capture its output
  run gcloud storage cat gs://$BUCKET/$TESTFILE
  assert_success

  # Calculate the actual length of the output
  local actual_length=$(echo "$output" | wc -c)

  assert_equal $actual_length $expected_length

  # # Check if the lengths match
  # if [ "$actual_length" -ne "$expected_length" ]; then
  #   printf "Expected length: %d, got %d\n" "$expected_length" "$actual_length"
  #   assert_failure #Incorrect byte length
  # fi
}

@test "Test gcloud storage rm" {
  run gcloud storage rm gs://$BUCKET/$TESTFILE
  assert_success
}

#TODO add more gcloud cat tests...